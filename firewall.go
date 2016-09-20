package clcv2

import (
	"fmt"
)

// IntraDataCenterFirewallPolicy represents an intra-datacenter firewall policy
type IntraDataCenterFirewallPolicy struct {
	// ID of the firewall policy
	ID string `json:"id"`

	// The state of the policy: either
	// - active  (policy is available and working as expected),
	// - error   (policy creation did not complete as expected) or
	// - pending (the policy is in the process of being created)
	Status string `json:"status"`

	// Indicates if the policy is enabled (true) or disabled (false)
	Enabled bool `json:"enabled"`

	// Source addresses for traffic on the originating firewall, specified using CIDR notation
	Source []string `json:"source"`

	// Destination addresses for traffic on the terminating firewall, specified using CIDR notation
	Destination []string `json:"destination"`

	// Short code for a particular account
	DestinationAccount string `json:"destinationAccount"`

	// Type of ports associated with the policy.
	// Supported ports include:
	// - any,
	// - icmp,
	// - TCP and UDP with single ports (tcp/123, udp/123) and
	// - port ranges (tcp/123-456, udp/123-456).
	// Some common ports include: tcp/21 (for FTP), tcp/990 (FTPS), tcp/80 (HTTP 80), tcp/8080 (HTTP 8080),
	//                            tcp/443 (HTTPS 443), icmp (PING), tcp/3389 (RDP), and tcp/22 (SSH/SFTP).
	Ports []string `json:"ports"`

	// Collection of entity links that point to resources related to this list of firewall policies
	Links []Link `json:"links"`
}

// Get details of a specific firewall policy associated with a given account in a given data center
// (an "intra data center firewall policy").
// @location: Short string representing the data center to query.
// @policyId: ID of the firewall policy to display.
func (c *Client) GetIntraDataCenterFirewallPolicy(location, policyId string) (res IntraDataCenterFirewallPolicy, err error) {
	path := fmt.Sprintf("/v2-experimental/firewallPolicies/%s/%s/%s", c.AccountAlias, location, policyId)
	err = c.getCLCResponse("GET", path, nil, &res)
	return
}

// GetIntraDataCenterFirewallPolicyList lists intra-datacenter firewall policies for the given account.
// Optionally filter results to policies associated with a second "destination" account.
// @location:   Short string representing the data center to query.
// @dstAccount: Optional destination account (empty string to omit).
func (c *Client) GetIntraDataCenterFirewallPolicyList(location, dstAccount string) (res []IntraDataCenterFirewallPolicy, err error) {
	path := fmt.Sprintf("/v2-experimental/firewallPolicies/%s/%s", c.AccountAlias, location)
	if dstAccount != "" {
		path += fmt.Sprintf("?destinationAccount=%s", dstAccount)
	}
	err = c.getCLCResponse("GET", path, nil, &res)
	return
}

/*
 * Cross-DataCenter Firewall Policies
 */
// CrossDataCenterFirewallPolicy represents an intra-datacenter firewall policy
type CrossDataCenterFirewallPolicy struct {
	// ID of the firewall policy
	ID string `json:"id"`

	// The state of the policy, possibly one of "active", "error", "pending"
	Status string `json:"status"`

	// Indicates if the policy is enabled
	Enabled bool `json:"enabled"`

	// Source network in CIDR notation
	SourceCIDR string `json:"sourceCidr"`

	// Source Account
	SourceAccount string `json:"sourceAccount"`

	// Source DataCenter Alias
	SourceLocation string `json:"sourceLocation"`

	// Destination network in CIDR notation
	DestCIDR string `json:"destinationCidr"`

	// Destination Account
	DestAccount string `json:"destinationAccount"`

	// Destination DataCenter Alias
	DestLocation string `json:"destinationLocation"`

	// Collection of entity links, listing e.g. verbs for self
	Links []Link
}

// GetCrossDataCenterFirewallPolicyList lists cross-datacenter firewall policies for the given account.
// Optionally filter results to policies associated with a second "destination" account.
// @location:   short string representing the data center to query
// @dstAccount: optional destination account (empty string to omit)
func (c *Client) GetCrossDataCenterFirewallPolicyList(location, dstAccount string) (res []CrossDataCenterFirewallPolicy, err error) {
	var path = fmt.Sprintf("/v2-experimental/crossDcFirewallPolicies/%s/%s", c.AccountAlias, location)

	if dstAccount != "" {
		path += fmt.Sprintf("?destinationAccount=%s", dstAccount)
	}
	err = c.getCLCResponse("GET", path, nil, &res)
	return
}

// GetCrossDataCenterFirewallPolicy returns details of a single cross-datacenter policy
// @location: data center location
// @id:       cross-datacenter policy ID
func (c *Client) GetCrossDataCenterFirewallPolicy(location, id string) (res CrossDataCenterFirewallPolicy, err error) {
	var path = fmt.Sprintf("/v2-experimental/crossDcFirewallPolicies/%s/%s/%s", c.AccountAlias, location, id)

	err = c.getCLCResponse("GET", path, nil, &res)
	return
}
