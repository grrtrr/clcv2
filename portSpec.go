package clcv2

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// PortSpecs implements the flag.Value interface for Port, allowing to use flags repeatedly.
type PortSpecs []Port

// PortSpecString is a Port to be used for other applications
type PortSpecString PortSpecs

// MarshalJSON to JSON
func (p PortSpecString) MarshalJSON() ([]byte, error) {
	var specs = make([]string, len(p))

	for i, port := range p {
		specs[i] = fmt.Sprintf("\"%s\"", port)
	}
	return []byte(fmt.Sprintf("[%s]", strings.Join(specs, ", "))), nil
}

// Port specifies a port or a port range.
// Note: the struct and json tags refer to the format used by the Public IP API.
//       To reuse the struct, convert it into string using Port.String()
type Port struct {
	// The specific protocol to support for the given port(s).
	// Should be either "tcp", "udp", or "icmp".
	Protocol string `json:"protocol"`

	// The port to open for the given protocol.
	// If defining a range of ports, this represents the first port in the range.
	Port int `json:"port"`

	// If defining a range of ports, optionally provide the last number of the range.
	PortTo int `json:"portTo"`
}

// String implements the Stringer interface for Port
func (p Port) String() string {
	if p.Protocol == "icmp" { // icmp does not have a port
		return p.Protocol
	}

	// Note: v2 documentation specifies lower case, reality and examples use upper case
	portSpec := fmt.Sprintf("%s/%d", strings.ToLower(p.Protocol), p.Port)
	if p.PortTo != 0 {
		portSpec += fmt.Sprintf("-%d", p.PortTo)
	}
	return portSpec
}

// String implements the flag.Value String method for PortSpecs
func (p PortSpecs) String() string {
	var specs = make([]string, len(p))

	for i, port := range p {
		specs[i] = fmt.Sprint(port)
	}
	return fmt.Sprintf("[%s]", strings.Join(specs, ", "))
}

// Type implements pflag.Value.Type
func (*PortSpecs) Type() string {
	return "CLCv2 Port Specifications"
}

// Set implements the flag.Value Set method for SourceRestriction.
func (p *PortSpecs) Set(val string) error {
	ps, err := ParsePortSpec(val)
	if err != nil {
		return errors.Errorf("invalid port specifier %q: %s", val, err)
	}
	*p = append(*p, ps)
	return nil
}

// ParsePortSpec parses a string satisfying one of the following formats into a Port:
// a) <proto>:<servicename>         - single port referenced via an /etc/services name
// b) <servicename>                 - abbreviation for tcp:<servicename>
//    <number>                      - abbreviation for tcp:<number>
// c) ping or icmp                  - an abbreviation for icmp:0
// c) <proto>/<portNumber>          - single numeric port
// d) <proto>/<startPort>-<endPort> - numeric port range
func ParsePortSpec(ps string) (p Port, err error) {
	var numericPorts = regexp.MustCompile(`^(\d+-)?\d+$`)
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
		return p, errors.Errorf("unsupported protocol type %q", proto)
	}

	if numericPorts.MatchString(el[1]) { /* numeric port or port range */
		if idx := strings.Index(el[1], "-"); idx > 0 {
			/* port range */
			if p.Port, err = strconv.Atoi(el[1][:idx]); err != nil || p.Port < 0 || p.Port > 65535 {
				return p, errors.Errorf("invalid start port %q", el[1][:idx])
			} else if p.PortTo, err = strconv.Atoi(el[1][idx+1:]); err != nil || p.PortTo < 0 || p.PortTo > 65535 {
				return p, errors.Errorf("invalid end port %q", el[1][idx+1:])
			} else if p.PortTo < p.Port {
				return p, errors.Errorf("invalid port range %q", el[1])
			}
		} else if p.Port, err = strconv.Atoi(el[1]); err == nil {
			/* numeric port */
			if p.Port < 0 || p.Port > 65535 {
				return p, errors.Errorf("invalid port %q", el[1])
			}
		}
	} else if p.Protocol != "udp" && p.Protocol != "tcp" {
		/* service name, as defined in /etc/services for UDP and TCP */
		return p, errors.Errorf("invalid service specification %q for %s", el[1], p.Protocol)
	} else {
		/* CLCv2 uses IPv4 addresses exclusively - look up v4 port names only */
		p.Port, err = net.LookupPort(p.Protocol+"4", el[1])
	}
	return
}
