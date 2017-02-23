package clcv2

import (
	"fmt"
	"net"
	"time"

	"github.com/pkg/errors"
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
// @location: The Network's home datacenter alias.
// @account:  (parent) AccountAlias to use (defaults to client's default AccountAlias)
func (c *Client) GetNetworks(location, account string) (nets []Network, err error) {
	if account == "" {
		account = c.AccountAlias
	}
	err = c.getCLCResponse("GET", fmt.Sprintf("/v2-experimental/networks/%s/%s", account, location), nil, &nets)
	return
}

// GetNetworkIdByName looks up a network by @name in @location
func (c *Client) GetNetworkIdByName(name, location string) (*Network, error) {
	return c.LookupMatchingNet(location, func(n *Network) bool { return n.Name == name })
}

// GetNetworkIdByCIDR looks up a network by @cidr in @location.
func (c *Client) GetNetworkIdByCIDR(cidr, location string) (*Network, error) {
	return c.LookupMatchingNet(location, func(n *Network) bool { return n.Cidr == cidr })
}

// GetNetworkIdByIP tries to find a network matching @ip in @location.
func (c *Client) GetNetworkIdByIP(ip, location string) (*Network, error) {
	ipAddr := net.ParseIP(ip)
	if ipAddr == nil {
		return nil, errors.Errorf("invalid IP address %s", ip)
	}
	return c.LookupMatchingNet(location, func(n *Network) bool {
		_, network, err := net.ParseCIDR(n.Cidr)
		if err != nil {
			panic(errors.Errorf("invalid CIDR address in %+v: %s", n, err))
		}
		return network.Contains(ipAddr)
	})
}

// LookupMatchingNet looks up a network in the scope of @c for which @matches returns true.
func (c *Client) LookupMatchingNet(location string, matches func(*Network) bool) (*Network, error) {
	// First pass: try the (lowest) scope of the given account alias
	nets, err := c.GetNetworks(location, c.AccountAlias)
	if err != nil {
		return nil, errors.Errorf("failed to lookup up %s networks in %s: %s",
			c.AccountAlias, location, err)
	}
	for idx := range nets {
		if matches(&nets[idx]) {
			return &nets[idx], nil
		}
	}
	// Second pass: check the parent account, if any
	if parentAcct := c.RegisteredAccountAlias(); parentAcct != c.AccountAlias {
		if nets, err = c.GetNetworks(location, parentAcct); err != nil {
			return nil, errors.Errorf("failed to lookup up %s networks in %s: %s",
				parentAcct, location, err)
		}
		for idx := range nets {
			if matches(&nets[idx]) {
				return &nets[idx], nil
			}
		}
	}
	return nil, nil
}

// Utility routine to look up a network by member IP @ip in @networks.
// Return pointer to matching network if found, else nil.
func NetworkByIP(ip string, networks []Network) (*Network, error) {
	ipAddr := net.ParseIP(ip)
	if ipAddr == nil {
		return nil, errors.Errorf("invalid IP address %s", ip)
	}
	for i := range networks {
		if _, net, err := net.ParseCIDR(networks[i].Cidr); err != nil {
			return nil, errors.Errorf("failed to parse CIDR %s: %s", networks[i].Cidr, err)
		} else if net.Contains(ipAddr) {
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

// NetworkDetails is returned in response to a get-network-details call.
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

	if networks, err := c.GetNetworks(location, c.AccountAlias); err != nil {
		return nil, err
	} else if len(networks) == 0 {
		return nil, errors.Errorf("No %s networks in %s available", c.AccountAlias, location)
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
			return nil, errors.Errorf("failed to query details of network %s: %s", id, err)
		}
		for idx := range details.IpAddresses {
			if details.IpAddresses[idx].Address == ip {
				/* Bail out if match is ambiguous, i.e. guarantee at most 1 match */
				if iad != nil {
					return nil, errors.Errorf("Duplicate match for %s: %+v and %+v\n",
						ip, iad, details.IpAddresses[idx])
				}
				iad = &details.IpAddresses[idx]
			}
		}
	}
	return
}

// claimNetworkStatus is returned by a GET on the URI returned from a claim-network POST operation.
type claimNetworkStatus struct {
	RequestType string
	Status      QueueStatus
	Source      struct {
		Username    string
		RequestedAt time.Time
	}
	Summary struct {
		BlueprintID uint64
		LocationID  string
		Links       []StatusLink // Contains a single link "network", which then contains the network ID
	}
}

// ClaimNetwork claims a new network in @datacentre and returns a status URI for this request.
func (c *Client) ClaimNetwork(datacentre string, cb func(QueueStatus)) (networkID string, err error) {
	var (
		path = fmt.Sprintf("/v2-experimental/networks/%s/%s/claim", c.AccountAlias, datacentre)
		res  struct {
			ID  string `json:"operationId"`
			URI string `json:"URI"`
		}
		cs claimNetworkStatus
	)

	if err := c.getCLCResponse("POST", path, nil, &res); err != nil {
		return "", err
	}

	for prevStatus := Unknown; ; {
		if err := c.getCLCResponse("GET", res.URI, nil, &cs); err != nil {
			return "", errors.Errorf("failed to query claim-network queue status: %s", err)
		}
		if cs.Status != prevStatus {
			if cb != nil {
				cb(cs.Status)
			}
			prevStatus = cs.Status
		}
		if cs.Status == Failed {
			return "", errors.Errorf("claim-network failed with status %q", cs.Status)
		} else if cs.Status == Succeeded {
			for _, link := range cs.Summary.Links {
				if link.Rel == "network" {
					return link.Id, nil
				}
			}
			return "", errors.Errorf("claim-network #%d succeeded, but returned no network ID", cs.Summary.BlueprintID)
		}
		time.Sleep(5 * time.Second) // operation may take several minutes
	}
}

// ReleaseNetwork releases @networkID in @datacentre
func (c *Client) ReleaseNetwork(datacentre, networkID string) error {
	path := fmt.Sprintf("/v2-experimental/networks/%s/%s/%s/release", c.AccountAlias, datacentre, networkID)
	return c.getCLCResponse("POST", path, nil, nil)
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
