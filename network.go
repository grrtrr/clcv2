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
	Id string

	// The network address, specified using CIDR notation
	Cidr string

	// Description of VLAN, a free text field that defaults to the VLAN number combined with the network address
	Description string

	// Gateway IP address of the network
	Gateway string

	// User-defined name of the network; the default is the VLAN number combined with the network address
	Name string

	// A screen of numbers used for routing traffic within a subnet
	Netmask string

	// Network type, usually private for networks created by the user
	Type string

	// Unique number assigned to the VLAN
	Vlan int

	// Collection of entity links that point to resources related to this list of networks
	Links []Link
}

// Get the list of networks available for a given account in a given data center.
// @location:  The Network's home datacenter alias.
func (c *Client) GetNetworks(location string) (nets []Network, err error) {
	err = c.getCLCResponse("GET", fmt.Sprintf("/v2-experimental/networks/%s/%s", c.AccountAlias, location), nil, &nets)
	return
}

// Get the network Id by the network name
// @name:      Name of the network to match.
// @location:  The Network's home datacenter alias.
func (c *Client) GetNetworkIdByName(name, location string) (net *Network, err error) {
	if nets, err := c.GetNetworks(location); err == nil {
		for idx := range nets {
			if nets[idx].Name == name {
				return &nets[idx], nil
			}
		}
	}
	return
}

// Get the network Id by CIDR
// @cidr:      CIDR of the network to match.
// @location:  The Network's home datacenter alias.
func (c *Client) GetNetworkIdByCIDR(cidr, location string) (net *Network, err error) {
	if nets, err := c.GetNetworks(location); err == nil {
		for idx := range nets {
			if nets[idx].Cidr == cidr {
				return &nets[idx], nil
			}
		}
	}
	return
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

type IpAddressDetails struct {
	// An IP Address on the Network
	Address string

	// Indicates claimed status of the address, either true or false
	Claimed bool

	// Indicates if the IP address is private, publicMapped, or virtual
	Type string

	// ID of the server associated with the IP address, if claimed
	// NOTE: this field is not always present (e.g. in CA2, it is sometimes missing).
	Server string
}

type NetworkDetails struct {
	Network
	IpAddresses []IpAddressDetails
}

// Get the details of a specific network in a given data center for a given account.
// @datacentre: Short string representing the data center you are querying.
// @network:    ID of the Network to query.
// @ipQuery:    Optional component of the query to request details of IP Addresses in a certain state.
//              Should be one of the following:
//              - "none"    (returns details of the network only),
//              - "claimed" (returns details of the network as well as information about claimed IP addresses),
//              - "free"    (returns details of the network as well as information about free IP addresses) or
//              - "all"     (returns details of the network as well as information about all IP addresses).
func (c *Client) GetNetworkDetails(datacentre, network, ipQuery string) (det NetworkDetails, err error) {
	path := fmt.Sprintf("/v2-experimental/networks/%s/%s/%s?ipAddresses=%s", c.AccountAlias, datacentre, network, ipQuery)
	err = c.getCLCResponse("GET", path, nil, &det)
	return
}

// Utility routine to extract only the list of IP addresses from @detailsList
func ExtractIPs(detailsList []IpAddressDetails) (ips []string) {
	for _, ip := range detailsList {
		ips = append(ips, ip.Address)
	}
	return
}

// Look up network details (server) given just an IP address.
// @ip:       Server IP address to look for
// @location: Location (data centre alias) to look @ip up in.
// Return details for @ip, nil if not found, or error.
func (c *Client) GetNetworkDetailsByIp(ip, location string) (iad *IpAddressDetails, err error) {
	var candidateNetworkIds []string

	if networks, err := c.GetNetworks(location); err != nil {
		return nil, err
	} else if len(networks) == 0 {
		return nil, fmt.Errorf("No %s networks in %s available", c.AccountAlias, location)
	} else {
		if net, err := NetworkByIP(ip, networks); err != nil {
			return nil, err
		} else if net != nil {
			candidateNetworkIds = []string{net.Id}
		} else { /* may be public address */
			for idx := range networks {
				candidateNetworkIds = append(candidateNetworkIds, networks[idx].Id)
			}
		}
	}

	for _, id := range candidateNetworkIds {
		details, err := c.GetNetworkDetails(location, id, "claimed")
		if err != nil {
			return nil, fmt.Errorf("failed to query details of network %s: %s", id, err)
		}
		for idx := range details.IpAddresses {
			if details.IpAddresses[idx].Address == ip {
				/* Bail out if match is ambiguous, i.e. guarantee at most 1 match */
				if iad != nil {
					return nil, fmt.Errorf("Duplicate match for %s: %+v and %+v\n",
						ip, iad, details.IpAddresses[idx])
				}
				iad = &details.IpAddresses[idx]
			}
		}
	}
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
	return c.getCLCResponse("PUT", path, &struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}{name, description}, nil)
}
