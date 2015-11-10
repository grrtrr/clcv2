package utils

import (
	"github.com/grrtrr/exit"
	"reflect"
	"sort"
	"fmt"
	"net"
)

/*
 * Sort IP addresses in ascending order
 */
type byIPv4Address []net.IP

func (ia byIPv4Address) Len() int {
	return len(ia)
}
func (ia byIPv4Address) Swap(i, j int) {
	(ia)[i], (ia)[j] = (ia)[j], (ia)[i]
}

func (ia byIPv4Address) Less(i, j int) bool {
	a,b := (ia)[i].To4(), (ia)[j].To4()
	if a[0] != b[0] {
		return a[0] < b[0]
	}
	if a[1] != b[1] {
		return a[1] < b[1]
	}
	if a[2] != b[2] {
		return a[2] < b[2]
	}
	return a[3] < b[3]
}

// Collapse an unsorted array @in of IPv4 addresses into sorted ranges in @out.
func CollapseIpRanges(in []string) (out []string) {
	var start, end, prev net.IP
	var ips []net.IP

	for _, ip_string := range in {
		ip := net.ParseIP(ip_string)
		if ip == nil {
			exit.Errorf("Invalid IP address %q", ip_string)
		}
		ip = ip.To4()
		if ip == nil {
			exit.Errorf("Not an IPv4 address: %q", ip_string)
		}
		ips = append(ips, ip)
	}
	sort.Stable(byIPv4Address(ips))

	for _, ip := range ips {
		if prev != nil && reflect.DeepEqual(ip[:3], prev[:3]) && ip[3] == prev[3]+1 {
			end = ip
		} else {
			if end != nil {
				out = append(out, fmt.Sprintf("%s-%d", start, end[3]))
				end = nil
			} else if prev != nil {
				out = append(out, prev.String())
			}
			start = ip
		}
		prev = ip
	}
	if end != nil {
		out = append(out, fmt.Sprintf("%s-%d", start, end[3]))
	}

	return out
}
