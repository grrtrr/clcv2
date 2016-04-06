package clcv2

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

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

// PortSpecs implenents the flag.Value interface for PublicPort
type PortSpecs []PublicPort

// PublicPort specifies a port mapping associated with a given PublicIPAddress.
type PublicPort struct {
	// The specific protocol to support for the given port(s).
	// Should be either "tcp", "udp", or "icmp".
	Protocol string `json:"protocol"`

	// The port to open for the given protocol.
	// If defining a range of ports, this represents the first port in the range.
	Port int `json:"port"`

	// If defining a range of ports, optionally provide the last number of the range.
	PortTo int `json:"portTo"`
}

// String implements the flag.Value String method for PortSpecs.
func (p *PortSpecs) String() string {
	var specs = make([]string, len(*p))

	for i := range *p {
		specs[i] = fmt.Sprint((*p)[i])
	}
	return strings.Join(specs, ", ")
}

// Set implements the flag.Value Set method for SourceRestriction.
func (p *PortSpecs) Set(val string) error {
	ps, err := ParsePortSpec(val)
	if err != nil {
		return fmt.Errorf("invalid port specifier %q: %s", val, err)
	}
	*p = append(*p, ps)
	return nil
}

// String implements the Stringer interface for PublicPort
func (p PublicPort) String() string {
	// Note: v2 documentation specifies lower case, reality and examples use upper case
	var portSpec = fmt.Sprintf("%s/%d", strings.ToLower(p.Protocol), p.Port)

	if p.PortTo != 0 {
		portSpec += fmt.Sprintf("-%d", p.PortTo)
	}
	return portSpec
}

// ParsePortSpec parses a string satisfying one of the following formats into a PublicPort:
// a) <proto>:<servicename>         - single port referenced via an /etc/services name
// b) <servicename>                 - abbreviation for tcp:<servicename>
//    <number>                      - abbreviation for tcp:<number>
// c) ping or icmp                  - an abbreviation for icmp:0
// c) <proto>/<portNumber>          - single numeric port
// d) <proto>/<startPort>-<endPort> - numeric port range
func ParsePortSpec(ps string) (p PublicPort, err error) {
	var el = strings.Split(ps, "/")

	if len(el) != 2 {
		if ps == "ping" || ps == "icmp" {
			p.Protocol = "icmp"
			return
		} else if ps == "rdp" {
			/* RDP may not be listed in /etc/services */
			p.Protocol, p.Port = "tcp", 3389
			return
		}
		/* Otherwise assume it's a tcp Service Name */
		return ParsePortSpec("tcp/" + ps)
	}

	switch proto := strings.ToLower(el[0]); proto {
	case "udp", "tcp", "icmp": /* only supported formats */
		p.Protocol = proto
	default:
		return p, fmt.Errorf("unsupported protocol type %q", proto)
	}

	if idx := strings.Index(el[1], "-"); idx > 0 {
		/* port range */
		if p.Port, err = strconv.Atoi(el[1][:idx]); err != nil || p.Port < 0 || p.Port > 65535 {
			return p, fmt.Errorf("invalid start port %q", el[1][:idx])
		} else if p.PortTo, err = strconv.Atoi(el[1][idx+1:]); err != nil || p.PortTo < 0 || p.PortTo > 65535 {
			return p, fmt.Errorf("invalid end port %q", el[1][idx+1:])
		} else if p.PortTo < p.Port {
			return p, fmt.Errorf("invalid port range %q", el[1])
		}
	} else if p.Port, err = strconv.Atoi(el[1]); err == nil {
		/* numeric port */
		if p.Port < 0 || p.Port > 65535 {
			return p, fmt.Errorf("invalid port %q", el[1])
		}
	} else {
		/* service name, as defined in /etc/services */
		if p.Protocol != "udp" && p.Protocol != "tcp" {
			return p, fmt.Errorf("invalid service specification %q for %s", el[1], p.Protocol)
		}
		/* CLCv2 uses IPv4 addresses exclusively - look up v4 port names only */
		p.Port, err = net.LookupPort(p.Protocol+"4", el[1])
	}
	return
}

// SrcRestrictions implements the flag.Value interface - for populating SourceCIDR fields.
type SrcRestrictions []SourceCIDR

// SourceCIDR wraps the IP range allowed to access a public IP, specified using CIDR notation.
type SourceCIDR struct {
	Cidr string `json:"cidr"`
}

// String implements the flag.Value String method for SrcRestrictions.
func (s *SrcRestrictions) String() string {
	var cidrs = make([]string, len(*s))

	for i := range *s {
		cidrs[i] = (*s)[i].Cidr
	}
	return fmt.Sprintf("[%s]", strings.Join(cidrs, ", "))
}

// Set implements the flag.Value Set method for SrcRestrictions.
func (s *SrcRestrictions) Set(val string) error {
	_, net, err := net.ParseCIDR(val)
	if err != nil {
		return fmt.Errorf("invalid source restriction format %q: %s", val, err)
	}
	*s = append(*s, SourceCIDR{net.String()})
	return nil
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
	err = c.getResponse("GET", path, nil, &res)
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
