package ucloud

import (
	pumem "github.com/ucloud/ucloud-sdk-go/private/services/umem"
	"github.com/ucloud/ucloud-sdk-go/services/cube"
	"github.com/ucloud/ucloud-sdk-go/services/ipsecvpn"
	"github.com/ucloud/ucloud-sdk-go/services/uaccount"
	"github.com/ucloud/ucloud-sdk-go/services/uads"

	"github.com/ucloud/ucloud-sdk-go/services/iam"
	"github.com/ucloud/ucloud-sdk-go/services/label"
	"github.com/ucloud/ucloud-sdk-go/services/udb"
	"github.com/ucloud/ucloud-sdk-go/services/udisk"
	"github.com/ucloud/ucloud-sdk-go/services/udpn"
	"github.com/ucloud/ucloud-sdk-go/services/ufile"
	"github.com/ucloud/ucloud-sdk-go/services/ufs"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/services/uk8s"
	"github.com/ucloud/ucloud-sdk-go/services/ulb"
	"github.com/ucloud/ucloud-sdk-go/services/umem"
	"github.com/ucloud/ucloud-sdk-go/services/unet"
	"github.com/ucloud/ucloud-sdk-go/services/uphost"
	"github.com/ucloud/ucloud-sdk-go/services/vpc"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

// UCloudClient is the ucloud openapi client
type UCloudClient struct {
	region    string
	projectId string

	config     *ucloud.Config
	credential *auth.Credential

	uhostconn      *uhost.UHostClient
	unetconn       *unet.UNetClient
	ulbconn        *ulb.ULBClient
	vpcconn        *vpc.VPCClient
	uaccountconn   *uaccount.UAccountClient
	udiskconn      *udisk.UDiskClient
	udpnconn       *udpn.UDPNClient
	udbconn        *udb.UDBClient
	umemconn       *umem.UMemClient
	ipsecvpnClient *ipsecvpn.IPSecVPNClient
	ufsconn        *ufs.UFSClient
	us3conn        *ufile.UFileClient
	cubeconn       *cube.CubeClient
	uk8sconn       *uk8s.UK8SClient
	uadsconn       *uads.UADSClient
	iamconn        *iam.IAMClient
	labelconn      *label.LabelClient
	uphostconn     *uphost.UPHostClient
	// private services
	pumemconn *pumem.UMemClient
	// client for generic invocation
	genericClient *ucloud.Client
}
