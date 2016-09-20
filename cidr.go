package clcv2

/*
 * Routines to specify/parse CIDR strings
 */
import (
	"fmt"
	"net"
	"strings"
)

// CIDRs implements the flag.Value interface, allowing to speciy multiple CIDR values.
type CIDRs []string

// String implements the flag.Value String method for CIDRs.
func (c CIDRs) String() string {
	var cidrs = make([]string, len(c))

	for i, cidr := range c {
		cidrs[i] = fmt.Sprint(cidr)
	}
	return fmt.Sprintf("[%s]", strings.Join(cidrs, ", "))
}

// Set implements the flag.Value Set method for CIDRs.
func (c *CIDRs) Set(val string) error {
	_, net, err := net.ParseCIDR(val)
	if err != nil {
		return fmt.Errorf("invalid CIDR format %q: %s", val, err)
	}
	*c = append(*c, net.String())
	return nil
}

// SrcRestrictions is analogous to CIDRs. A separate implementation is needed, since it has a different type.
type SrcRestrictions []SourceCIDR

// SourceCIDR wraps the IP range allowed to access a public IP, specified using CIDR notation.
type SourceCIDR struct {
	Cidr string `json:"cidr"`
}

// String implements the flag.Value String method for SrcRestrictions.
func (s SrcRestrictions) String() string {
	var cidrs = make([]string, len(s))

	for i, cidr := range s {
		cidrs[i] = cidr.Cidr
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
