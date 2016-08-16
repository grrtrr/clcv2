package clcv2

import (
	"fmt"
)

type FWPolicy struct {
	// ID of the firewall policy
	Id			string

	// The state of the policy: either
	// - active  (policy is available and working as expected),
	// - error   (policy creation did not complete as expected) or
	// - pending (the policy is in the process of being created)
	Status			string

	// Indicates if the policy is enabled (true) or disabled (false)
	Enabled			bool

	// Source addresses for traffic on the originating firewall, specified using CIDR notation
	Source			[]string

	// Destination addresses for traffic on the terminating firewall, specified using CIDR notation
	Destination		[]string

	// Short code for a particular account
	DestinationAccount	string

	// Type of ports associated with the policy.
	// Supported ports include:
	// - any,
	// - icmp,
	// - TCP and UDP with single ports (tcp/123, udp/123) and
	// - port ranges (tcp/123-456, udp/123-456).
	// Some common ports include: tcp/21 (for FTP), tcp/990 (FTPS), tcp/80 (HTTP 80), tcp/8080 (HTTP 8080),
	//                            tcp/443 (HTTPS 443), icmp (PING), tcp/3389 (RDP), and tcp/22 (SSH/SFTP).
	Ports			[]string

	// Collection of entity links that point to resources related to this list of firewall policies
	Links			[]Link

}

// Get details of a specific firewall policy associated with a given account in a given data center
// (an "intra data center firewall policy").
// @location: Short string representing the data center to query.
// @policyId: ID of the firewall policy to display.
func (c *Client) GetFWPolicy(location, policyId string) (res FWPolicy, err error) {
	path := fmt.Sprintf("/v2-experimental/firewallPolicies/%s/%s/%s", c.AccountAlias, location, policyId)
	err = c.getCLCResponse("GET", path, nil, &res)
	return
}

// List firewall policies associated with a given account in a given data center
// ("intra data center firewall policies").
// Optionally filter results to policies associated with a second "destination" account.
// @location:   Short string representing the data center to query.
// @dstAccount: Optional destination account (empty string to omit).
func (c *Client) GetFWPolicyList(location, dstAccount string) (res []FWPolicy, err error) {
	path := fmt.Sprintf("/v2-experimental/firewallPolicies/%s/%s", c.AccountAlias, location)
	if dstAccount != "" {
		path += fmt.Sprintf("?destinationAccount=%s", dstAccount)
	}
	err = c.getCLCResponse("GET", path, nil, &res)
	return
}
