package clcv2

import (
	"fmt"
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
