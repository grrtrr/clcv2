package clcv2

import (
	"fmt"
	"net"
)

/*
 * Network Listing
 */
type Network struct {
	// ID of the network
	Id		string

	// The network address, specified using CIDR notation
	Cidr		string

	// Description of VLAN, a free text field that defaults to the VLAN number combined with the network address
	Description	string

	// Gateway IP address of the network
	Gateway		string

	// User-defined name of the network; the default is the VLAN number combined with the network address
	Name		string

	// A screen of numbers used for routing traffic within a subnet
	Netmask		string

	// Network type, usually private for networks created by the user
	Type		string

	// Unique number assigned to the VLAN
	Vlan		int

	// Collection of entity links that point to resources related to this list of networks
	Links		[]Link
}

// Get the list of networks available for a given account in a given data center.
// @location:  The Network's home datacenter alias.
func (c *Client) GetNetworks(location string) (nets []Network, err error) {
	err = c.getResponse("GET", fmt.Sprintf("/v2-experimental/networks/%s/%s", c.AccountAlias, location), nil, &nets)
	return
}

type IpAddressDetails struct {
	// An IP Address on the Network
	Address	string

	// Indicates claimed status of the address, either true or false
	Claimed	bool

	// ID of the server associated with the IP address, if claimed
	Server	string

	// Indicates if the IP address is private, publicMapped, or virtual
	Type	string
}

type NetworkDetails struct {
	Network
	IpAddresses	[]IpAddressDetails
}

// Get the details of a specific network in a given data center for a given account.
// @datacentre: Short string representing the data center you are querying.
// @network:    ID of the Network to query.
// @ipQuery:    Optional component of the query to request details of IP Addresses in a certain state.
//              Should be one of the following:
//              - "none" (returns details of the network only),
//              - "claimed" (returns details of the network as well as information about claimed IP addresses),
//              - "free" (returns details of the network as well as information about free IP addresses) or
//              - "all" (returns details of the network as well as information about all IP addresses).
func (c *Client) GetNetworkDetails(datacentre, network, ipQuery string) (det NetworkDetails, err error) {
	path := fmt.Sprintf("/v2-experimental/networks/%s/%s/%s?ipAddresses=%s", c.AccountAlias, datacentre, network, ipQuery)
	err = c.getResponse("GET", path, nil, &det)
	return
}

// Update the attributes of a given Network via PUT.
// @datacentre:  Short string representing the data center hosting @network.
// @network:     ID of the Network to update.
// @name:        User-defined name of the network
//               (the default is the VLAN number combined with the network address).
// @description: Description of VLAN, a free text field that defaults to the VLAN number plus network address.
func (c *Client) UpdateNetwork(datacentre, network, name, description string) error {
	path := fmt.Sprintf("/v2-experimental/networks/%s/%s/%s", c.AccountAlias, datacentre, network)
	return c.getResponse("PUT", path, &struct{
		Name		string	`json:"name"`
		Description	string	`json:"description"`
	} { name, description }, nil)
}

// Utility routine to look up a network by member IP @ips in @networks.
// Return pointer to matching network if found, else nil.
func NetworkByIP(ips string, networks []Network) (*Network, error) {
	ip := net.ParseIP(ips)
	if ip == nil {
		return nil, fmt.Errorf("Invalid IP address %s", ips)
	}
	for i := range networks {
		if _, net, err := net.ParseCIDR(networks[i].Cidr); err != nil {
			return nil, fmt.Errorf("Failed to parse CIDR %s: %s", networks[i].Cidr, err)
		} else if net.Contains(ip) {
			return &networks[i], nil
		}
	}
	return nil, nil
}
