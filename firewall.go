package clcv2

import (
	"fmt"
	"path"
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
	var path = fmt.Sprintf("/v2-experimental/firewallPolicies/%s/%s/%s", c.credentials.AccountAlias, location, policyId)
	err = c.getCLCResponse("GET", path, nil, &res)
	return res, err
}

// GetIntraDataCenterFirewallPolicyList lists intra-datacenter firewall policies for the given account.
// Optionally filter results to policies associated with a second "destination" account.
// @location:   Short string representing the data center to query.
// @dstAccount: Optional destination account (empty string to omit).
func (c *Client) GetIntraDataCenterFirewallPolicyList(location, dstAccount string) (res []IntraDataCenterFirewallPolicy, err error) {
	var path = fmt.Sprintf("/v2-experimental/firewallPolicies/%s/%s", c.credentials.AccountAlias, location)
	if dstAccount != "" {
		path += fmt.Sprintf("?destinationAccount=%s", dstAccount)
	}
	err = c.getCLCResponse("GET", path, nil, &res)
	return res, err
}

// IntraDataCenterFirewallPolicyReq contains the requisite data to request a new cross-datacenter firewall policy.
type IntraDataCenterFirewallPolicyReq struct {
	// Source networks for traffic on the originating firewall, in CIDR notation.
	SourceCIDR CIDRs `json:"source"`

	// Destination networks for traffic on the terminating firewall, in CIDR notation.
	DestCIDR CIDRs `json:"destination"`

	// Destination Account (short code)
	DestAccount string `json:"destinationAccount"`

	// Type of ports associated with the policy.
	Ports PortSpecString `json:"ports"`
}

// CreateIntraDataCenterFirewallPolicy creates a new intra-datacenter firewall policy at @location.
func (c *Client) CreateIntraDataCenterFirewallPolicy(location string, req *IntraDataCenterFirewallPolicyReq) (id string, err error) {
	var res struct {
		Links []Link
	}

	err = c.getCLCResponse("POST", fmt.Sprintf("/v2-experimental/firewallPolicies/%s/%s", c.AccountAlias, location), req, &res)
	if err != nil {
		/* response failed */
	} else if link, err := extractLink(res.Links, "self"); err != nil {
		/* failed to extract link */
	} else {
		/* The ID is contained as the last path field of the Href URL. There is no separate ID field. */
		id = path.Base(link.Href)
	}
	return id, err
}

// DeleteIntraDataCenterFirewallPolicy deletes the given cross-datacenter firewall policy @id in datacenter @location.
func (c *Client) DeleteIntraDataCenterFirewallPolicy(location, id string) error {
	return c.getCLCResponse("DELETE", fmt.Sprintf("/v2-experimental/firewallPolicies/%s/%s/%s", c.AccountAlias, location, id), nil, nil)
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
	var path = fmt.Sprintf("/v2-experimental/crossDcFirewallPolicies/%s/%s", c.credentials.AccountAlias, location)

	if dstAccount != "" {
		path += fmt.Sprintf("?destinationAccount=%s", dstAccount)
	}
	err = c.getCLCResponse("GET", path, nil, &res)
	return res, err
}

// GetCrossDataCenterFirewallPolicy returns details of a single cross-datacenter policy
// @location: data center location
// @id:       cross-datacenter policy ID
func (c *Client) GetCrossDataCenterFirewallPolicy(location, id string) (res CrossDataCenterFirewallPolicy, err error) {
	var path = fmt.Sprintf("/v2-experimental/crossDcFirewallPolicies/%s/%s/%s", c.credentials.AccountAlias, location, id)

	err = c.getCLCResponse("GET", path, nil, &res)
	return res, err
}

// ToggleCrossDataCenterFirewallPolicy enables/disables the given cross-datacenter firewall policy.
// @location: data center location
// @id:       cross-datacenter policy ID
// @enable:   whether to enable or disable @id
func (c *Client) ToggleCrossDataCenterFirewallPolicy(location, id string, enable bool) error {
	var path = fmt.Sprintf("/v2-experimental/crossDcFirewallPolicies/%s/%s/%s?enabled=%t", c.AccountAlias, location, id, enable)

	return c.getCLCResponse("PUT", path, nil, nil) // the response is an empty "204 No Content"
}

// CrossDataCenterFirewallPolicyReq contains the requisite data to request a new cross-datacenter firewall policy.
type CrossDataCenterFirewallPolicyReq struct {
	// Source network in CIDR notation
	SourceCIDR string `json:"sourceCidr"`

	// Destination network in CIDR notation
	DestCIDR string `json:"destinationCidr"`

	// Destination Account (short code)
	DestAccount string `json:"destinationAccountId"`

	// Destination DataCenter Alias
	DestLocation string `json:"destinationLocationId"`

	// Whether the policy is enabled
	Enabled bool `json:"enabled"`
}

// CreateCrossDataCenterFirewallPolicy creates a new cross-datacenter firewall policy at @location.
func (c *Client) CreateCrossDataCenterFirewallPolicy(location string, req *CrossDataCenterFirewallPolicyReq) (res CrossDataCenterFirewallPolicy, err error) {
	err = c.getCLCResponse("POST", fmt.Sprintf("/v2-experimental/crossDcFirewallPolicies/%s/%s", c.AccountAlias, location), req, &res)
	return res, err
}

// DeleteCrossDataCenterFirewallPolicy deletes the given cross-datacenter firewall policy @id in datacenter @location.
func (c *Client) DeleteCrossDataCenterFirewallPolicy(location, id string) error {
	return c.getCLCResponse("DELETE", fmt.Sprintf("/v2-experimental/crossDcFirewallPolicies/%s/%s/%s", c.AccountAlias, location, id), nil, nil)
}
