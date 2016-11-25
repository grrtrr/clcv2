package clcv2

/*
 * Site-to-Site VPNs
 */
import "fmt"

// SiteToSiteVPN represents CLCv2 Site-to-Site VPN information.
type SiteToSiteVPN struct {
	// The VPN ID
	ID string `json:"id"`

	// Associated account alias
	AccountAlias string `json:"accountAlias"`

	// VPN creation/modification details
	ChangeInfo `json:"changeInfo"`

	// The various entities of this VPN:
	Local  LocalEntity  `json:"local"`
	Remote RemoteEntity `json:"remote"`
	IKE    IkeEntity    `json:"ike"`
	IPsec  IPsecEntity  `json:"ipsec"`

	// Self-referential links
	Links []Link `json:"links"`
}

// LocalEntity represents a local network.
type LocalEntity struct {
	// Local IP address for Site to Site VPN
	Address string `json:"address"`

	// Short code for a particular location
	LocationAlias string `json:"locationAlias"`

	// Friendly description for @LocationAlias
	LocationDesc string `json:"locationDescription"`

	// List of local subnets, specified using CIDR notation
	Subnets []string `json:"subnets"`
}

// RemoteEntity represents a remote network.
type RemoteEntity struct {
	// Remote IP address for Site to Site VPN
	Address string `json:"address"`

	// Friendly name of the site
	SiteName string `json:siteName"`

	// Friendly name of the device type
	DeviceType string `json:"deviceType"`

	// List of remote subnets, specified using CIDR notation
	Subnets []string `json:"subnets"`
}

// IkeEntity represents IKE VPN configuration details.
type IkeEntity struct {
	// Encryption algorithm, one of: aes128, aes192, aes256, tripleDES
	Encryption string `json:"encryption"`

	//  Hashing algorithm, one of: sha1_96, sha1_256, md5
	Hashing string `json:"hashing"`

	// One of:
	// - "group1" (legacy),
	// - "group2", or
	// - "group5"
	//If using AES with a cipher strength greater than 128-bit, or SHA2 for hashing, we recommend "group5",
	// otherwise "group2" is sufficient.
	DiffieHellmanGroup string `json:"diffieHellmanGroup"`

	// The pre-shared key is a shared secret that secures the VPN tunnel.
	// This value must be identical on both ends of the connection
	PreSharedKey string `json:"preSharedKey,omitempty"`

	// Lifetime in seconds (valid: 3600, 28800, 86400).
	// Lifetime is set to 28800 (8 hours) for IKE. This is not required to match, as the negotiation
	// will choose the shortest value supplied by either peer.
	Lifetime uint64 `json:"lifetime"`

	// Protocol mode, one of: main, aggressive
	Mode string `json:"mode"`

	// Specify if you wish this enabled or disabled.
	// Check your device defaults; for example, Cisco ASA defaults to 'on' (i.e. true),
	// while Netscreen/Juniper SSG or Juniper SRX default to 'off'. Our default is 'off' (i.e. false).
	DeadPeerDetection bool `json:"deadPeerDetection"`

	// NAT-Traversal: Allows connections to VPN end-points behind a NAT device.
	// Defaults to false. If you require NAT-T, you also need to provide the private IP address
	// that your VPN endpoint will use to identify itself.
	NatTraversal bool `json:"natTraversal"`
}

// IPsecEntity represents IPsec VPN configuration details.
type IPsecEntity struct {
	// Encryption algorithm, one of: aes128, aes192, aes256, tripleDES
	Encryption string `json:"encryption"`

	// Hashing algorithm, one of: sha1_96, sha1_256, md5
	Hashing string `json:"hashing"`

	// IPSec protocol, one of: esp, ah
	Protocol string `json:"protocol"`

	// PFS enabled or disabled, one of: disabled, group1, group2, group5
	// (we suggest enabled, using "group2", though "group5" is recommended with SHA2 hashing or AES-192 or AES-256)
	Pfs string `json:"pfs"`

	// Lifetime in seconds, e.g. 3600, 28800, 86400
	// This setting is not required to match, as the negotiation process will choose the shortest value supplied by either peer.
	Lifetime uint64 `json:"lifetime"`
}

// GetVPNs returns the list of site-to-site VPNs associated with the given client AccountAlias.
func (c *Client) GetVPNs() (res []SiteToSiteVPN, err error) {
	err = c.getCLCResponse("GET", fmt.Sprintf("/v2/siteToSiteVpn?account=%s", c.AccountAlias), nil, &res)
	return
}

// GetVPN returns details of the specified Site-to-Site VPN.
func (c *Client) GetVPN(vpnID string) (res SiteToSiteVPN, err error) {
	err = c.getCLCResponse("GET", fmt.Sprintf("/v2/siteToSiteVpn/%s?account=%s", vpnID, c.credentials.AccountAlias), nil, &res)
	return
}
