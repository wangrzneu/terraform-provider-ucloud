package ucloud

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/ucloud/ucloud-sdk-go/services/iam"
	"github.com/ucloud/ucloud-sdk-go/services/label"
	"github.com/ucloud/ucloud-sdk-go/services/uads"
	"github.com/ucloud/ucloud-sdk-go/services/uphost"

	"github.com/ucloud/ucloud-sdk-go/services/cube"
	"github.com/ucloud/ucloud-sdk-go/services/ufile"
	"github.com/ucloud/ucloud-sdk-go/services/ufs"
	"github.com/ucloud/ucloud-sdk-go/services/uk8s"

	"github.com/ucloud/ucloud-sdk-go/services/udpn"

	"github.com/ucloud/ucloud-sdk-go/external"
	"github.com/ucloud/ucloud-sdk-go/private/protocol/http"
	pumem "github.com/ucloud/ucloud-sdk-go/private/services/umem"
	"github.com/ucloud/ucloud-sdk-go/services/ipsecvpn"
	"github.com/ucloud/ucloud-sdk-go/services/sts"
	"github.com/ucloud/ucloud-sdk-go/services/uaccount"
	"github.com/ucloud/ucloud-sdk-go/services/udb"
	"github.com/ucloud/ucloud-sdk-go/services/udisk"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/services/ulb"
	"github.com/ucloud/ucloud-sdk-go/services/umem"
	"github.com/ucloud/ucloud-sdk-go/services/unet"
	"github.com/ucloud/ucloud-sdk-go/services/vpc"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
	"github.com/ucloud/ucloud-sdk-go/ucloud/log"
)

// Config is the configuration of ucloud meta data
type Config struct {
	PublicKey             string
	PrivateKey            string
	Profile               string
	SharedCredentialsFile string
	AssumeRole            *AssumeRoleConfig
	Region                string
	ProjectId             string
	Insecure              bool
	BaseURL               string
	MaxRetries            int
}

// AssumeRoleConfig is the configuration of assume role
type AssumeRoleConfig struct {
	Duration    time.Duration
	RoleURN     string
	Policy      string
	SessionName string
}

type cloudShellCredential struct {
	Cookie    string `json:"cookie"`
	Profile   string `json:"profile"`
	CSRFToken string `json:"csrf_token"`
}

// Client will returns a client with connections for all product
func (c *Config) Client() (*UCloudClient, error) {
	var client UCloudClient

	client.region = c.Region
	client.projectId = c.ProjectId

	cfg := ucloud.NewConfig()

	// set common attributes (region, project id, etc ...)
	cfg.Region = c.Region
	cfg.ProjectId = c.ProjectId

	// enable auto retry with http/connection error
	cfg.MaxRetries = c.MaxRetries
	cfg.LogLevel = log.PanicLevel
	cfg.UserAgent = "Terraform-UCloud/1.38.3"
	cfg.BaseUrl = c.BaseURL

	cred := auth.NewCredential()

	if isAcc() {
		//set DebugLevel for acceptance test
		cfg.LogLevel = log.DebugLevel

		// excepted logging
		cfg.SetActionLevel("GetRegion", log.WarnLevel)
	}

	var cloudShellCredHandler ucloud.HttpRequestHandler
	if len(c.Profile) > 0 {
		// load public/private key from shared credential file
		credV, err := external.LoadUCloudCredentialFile(c.SharedCredentialsFile, c.Profile)
		if err != nil {
			return nil, fmt.Errorf("cannot load shared %q credential file, %s", c.Profile, err)
		}
		cred = *credV
	} else if len(c.PublicKey) > 0 && len(c.PrivateKey) > 0 {
		// load public/private key from shared credential file
		cred.PublicKey = c.PublicKey
		cred.PrivateKey = c.PrivateKey
	} else if v := os.Getenv("CLOUD_SHELL"); v == "true" {
		csCred := make([]cloudShellCredential, 0)
		// load credential from default cloud shell credential path
		if err := loadJSONFile(defaultCloudShellCredPath(), &csCred); err != nil {
			return nil, fmt.Errorf("must set credential about public_key and private_key, %s", err)
		}
		// get default cloud shell credential
		defaultCsCred := &cloudShellCredential{}
		for i := 0; i < len(csCred); i++ {
			if csCred[i].Profile == "default" {
				defaultCsCred = &csCred[i]
				break
			}
		}
		if defaultCsCred == nil || len(defaultCsCred.Cookie) == 0 || len(defaultCsCred.CSRFToken) == 0 {
			return nil, fmt.Errorf("must set credential about public_key and private_key, default credential is null")
		}

		// set cloud shell client handler
		cloudShellCredHandler = func(c *ucloud.Client, req *http.HttpRequest) (*http.HttpRequest, error) {
			req.SetHeader("Cookie", defaultCsCred.Cookie)
			req.SetHeader("Csrf-Token", defaultCsCred.CSRFToken)
			return req, nil
		}
	} else {
		return nil, fmt.Errorf("must set credential about public_key and private_key")
	}
	if c.AssumeRole != nil {
		// get STS credential
		stsCredential, err := getSTSCredential(*c.AssumeRole, cfg, cred)
		if err != nil {
			return nil, fmt.Errorf("fail to get STS credential, %w", err)
		}
		cred = *stsCredential
	}
	// initialize client connections
	client.unetconn = unet.NewClient(&cfg, &cred)
	client.ulbconn = ulb.NewClient(&cfg, &cred)
	client.vpcconn = vpc.NewClient(&cfg, &cred)
	client.uaccountconn = uaccount.NewClient(&cfg, &cred)
	client.udiskconn = udisk.NewClient(&cfg, &cred)
	client.umemconn = umem.NewClient(&cfg, &cred)
	client.ipsecvpnClient = ipsecvpn.NewClient(&cfg, &cred)
	client.ufsconn = ufs.NewClient(&cfg, &cred)
	client.us3conn = ufile.NewClient(&cfg, &cred)
	client.cubeconn = cube.NewClient(&cfg, &cred)
	client.uadsconn = uads.NewClient(&cfg, &cred)
	client.iamconn = iam.NewClient(&cfg, &cred)
	client.labelconn = label.NewClient(&cfg, &cred)

	// initialize client connections for private usage
	client.pumemconn = pumem.NewClient(&cfg, &cred)

	longtimeCfg := cfg
	longtimeCfg.Timeout = 60 * time.Second
	client.udbconn = udb.NewClient(&longtimeCfg, &cred)
	client.uhostconn = uhost.NewClient(&longtimeCfg, &cred)
	client.udpnconn = udpn.NewClient(&longtimeCfg, &cred)
	client.uk8sconn = uk8s.NewClient(&longtimeCfg, &cred)
	client.uphostconn = uphost.NewClient(&longtimeCfg, &cred)
	client.genericClient = ucloud.NewClient(&longtimeCfg, &cred)
	if cloudShellCredHandler != nil {
		client.uhostconn.AddHttpRequestHandler(cloudShellCredHandler)
		client.unetconn.AddHttpRequestHandler(cloudShellCredHandler)
		client.ulbconn.AddHttpRequestHandler(cloudShellCredHandler)
		client.vpcconn.AddHttpRequestHandler(cloudShellCredHandler)
		client.uaccountconn.AddHttpRequestHandler(cloudShellCredHandler)
		client.udiskconn.AddHttpRequestHandler(cloudShellCredHandler)
		client.udpnconn.AddHttpRequestHandler(cloudShellCredHandler)
		client.udbconn.AddHttpRequestHandler(cloudShellCredHandler)
		client.umemconn.AddHttpRequestHandler(cloudShellCredHandler)
		client.ipsecvpnClient.AddHttpRequestHandler(cloudShellCredHandler)
		client.pumemconn.AddHttpRequestHandler(cloudShellCredHandler)
		client.ufsconn.AddHttpRequestHandler(cloudShellCredHandler)
		client.us3conn.AddHttpRequestHandler(cloudShellCredHandler)
		client.cubeconn.AddHttpRequestHandler(cloudShellCredHandler)
		client.uk8sconn.AddHttpRequestHandler(cloudShellCredHandler)
		client.uadsconn.AddHttpRequestHandler(cloudShellCredHandler)
		client.iamconn.AddHttpRequestHandler(cloudShellCredHandler)
		client.labelconn.AddHttpRequestHandler(cloudShellCredHandler)
		client.uphostconn.AddHttpRequestHandler(cloudShellCredHandler)
		client.genericClient.AddHttpRequestHandler(cloudShellCredHandler)
	}

	client.config = &cfg
	client.credential = &cred
	return &client, nil
}

func defaultCloudShellCredPath() string {
	return filepath.Join(userHomeDir(), ".ucloud", "credential.json")
}

func loadJSONFile(path string, p interface{}) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	c, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	err = json.Unmarshal(c, p)
	if err != nil {
		return err
	}
	return nil
}

func userHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

func getSTSCredential(assumeRole AssumeRoleConfig, config ucloud.Config, credential auth.Credential) (*auth.Credential, error) {
	// get STS credential
	stsClient := sts.NewClient(&config, &credential)
	var assumeRoleRequest sts.AssumeRoleRequest
	assumeRoleRequest.Policy = ucloud.String(assumeRole.Policy)
	assumeRoleRequest.RoleUrn = ucloud.String(assumeRole.RoleURN)
	assumeRoleRequest.RoleSessionName = ucloud.String(assumeRole.SessionName)
	assumeRoleRequest.DurationSeconds = ucloud.Int(int(assumeRole.Duration.Seconds()))
	assumeRoleResponse, err := stsClient.AssumeRole(&assumeRoleRequest)
	if err != nil {
		return nil, fmt.Errorf("fail to assume role, %w", err)
	}
	// set STS credential
	var stsCredential auth.Credential
	stsCredential.PublicKey = assumeRoleResponse.Credentials.AccessKeyId
	stsCredential.PrivateKey = assumeRoleResponse.Credentials.AccessKeySecret
	stsCredential.SecurityToken = assumeRoleResponse.Credentials.SecurityToken
	stsCredential.CanExpire = true
	expireTime, err := time.Parse(time.RFC3339, assumeRoleResponse.Credentials.Expiration)
	if err != nil {
		return nil, fmt.Errorf("fail to parse expiration time, %w", err)
	}
	stsCredential.Expires = expireTime
	return &stsCredential, nil
}
