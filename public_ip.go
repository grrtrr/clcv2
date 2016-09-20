package clcv2

import "fmt"

/*
 * Management of Public IP Addresses
 */
type PublicIPAddress struct {
	// The internal (private) IP address to map to the new public IP address.
	// If not provided, one will be assigned for you.
	InternalIPAddress string `json:"internalIPAddress,omitempty"`

	// The set of ports and protocols to allow access to for the new public IP address.
	// Only these specified ports on the respective protocols will be accessible when
	// accessing the server using the public IP address claimed here.
	Ports PortSpecs `json:"ports"`

	// The source IP address range allowed to access the new public IP address.
	// Used to restrict access to only the specified range of source IPs.
	SourceRestrictions SrcRestrictions `json:"sourceRestrictions"`
}

// Claim a public IP address and associate it with a server, allowing access to it on a given set of
// protocols and ports. It may also be set to restrict access based on a source IP range.
// @serverId: ID of the server to change.
func (c *Client) AddPublicIPAddress(serverId string, req *PublicIPAddress) (statusId string, err error) {
	return c.getStatus("POST", fmt.Sprintf("/v2/servers/%s/%s/publicIPAddresses", c.AccountAlias, serverId), req)
}

// Get the details for the public IP address of a server.
// @serverId: ID of the server to query.
// @publicIp: The specific public IP to return details about.
func (c *Client) GetPublicIPAddress(serverId, publicIp string) (res PublicIPAddress, err error) {
	path := fmt.Sprintf("/v2/servers/%s/%s/publicIPAddresses/%s", c.AccountAlias, serverId, publicIp)
	err = c.getCLCResponse("GET", path, nil, &res)
	return
}

// Update a public IP address on an existing server.
// @serverId: ID of the server to update.
// @publicIp: The specific public IP to return details about.
func (c *Client) UpdatePublicIPAddress(serverId, publicIp string, req *PublicIPAddress) (statusId string, err error) {
	path := fmt.Sprintf("/v2/servers/%s/%s/publicIPAddresses/%s", c.AccountAlias, serverId, publicIp)
	return c.getStatus("PUT", path, req)
}

// Release the given public IP address of a server so that it is no longer associated
// with the server and available to be claimed again by another server.
// @serverId: ID of the server to query.
// @publicIp: The specific public IP to return details about.
func (c *Client) RemovePublicIPAddress(serverId, publicIp string) (statusId string, err error) {
	path := fmt.Sprintf("/v2/servers/%s/%s/publicIPAddresses/%s", c.AccountAlias, serverId, publicIp)
	return c.getStatus("DELETE", path, nil)
}
