package cmd

import (
	"encoding/hex"
	"fmt"
	"net"

	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	var nic = &cobra.Command{
		Use:   "nic",
		Short: "Manage server NICs",
		Long:  "Add or remove server secondary network interface",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Both commands take <serverName> <net ID | name | CIDR>
			if len(args) < 2 {
				return errors.Errorf("Need a server name and a network ID or CIDR  for the secondary NIC")
			}
			setLocationBasedOnServerName(args[0])
			return nil
		},
	}

	nic.AddCommand(addNIC)
	nic.AddCommand(removeNIC)

	Root.AddCommand(nic)
}

// addNICFlags
var addNICFlags struct {
	ip string // IP address to assign to the secondary NIC
}

// resolveNet attempts to resolve @s into a hexadecimal ID of a network in @location.
// It supports hex ID, CIDR, IP address, or network name.
// NOTE: requires global @client to be initialized
func resolveNet(s, location string) (string, error) {
	var netw *clcv2.Network

	if _, err := hex.DecodeString(s); err == nil {
		/* already looks like a HEX ID */
		return s, nil
	} else if _, network, err := net.ParseCIDR(s); err == nil { // CIDR string
		if netw, err = client.GetNetworkIdByCIDR(s, location); err != nil {
			return "", errors.Errorf("failed to look up CIDR %q in %s: %s",
				network, location, err)
		}
	} else if ip := net.ParseIP(s); ip != nil { // IP address (without CIDR netmask)
		if netw, err = client.GetNetworkIdByIP(s, location); err != nil {
			return "", errors.Errorf("failed to look up IP %q in %s: %s",
				ip, location, err)
		}
	} else { // network name
		if netw, err = client.GetNetworkIdByName(s, location); err != nil {
			return "", errors.Errorf("failed to look up network %q in %s: %s",
				s, location, err)
		}
	}
	if netw == nil {
		return "", errors.Errorf("no network matching %q found in %s", s, location)
	}
	return netw.Id, nil
}

var addNIC = &cobra.Command{
	Use:   "add  <serverName>  <net (ID | CIDR | IP)>",
	Short: "Add a secondary NIC to server",
	Long:  "Add a secondary NIC to @server on network @net (using network ID, CIDR, or IP)",
	RunE: func(cmd *cobra.Command, args []string) error {
		var server, netStr = args[0], args[1]

		netID, err := resolveNet(netStr, location)
		if err != nil {
			exit.Errorf("failed to resolve %s: %s", netStr, err)
		}
		fmt.Println(netID, debug)
		return nil // XXX

		if err := client.ServerAddNic(server, netStr, addNICFlags.ip); err != nil {
			exit.Fatalf("failed to add NIC to %s: %s", server, err)
		}
		return nil
	},
}

var removeNIC = &cobra.Command{
	Use:     "remove  <serverName>  <net (ID | CIDR | IP)>",
	Aliases: []string{"rm", "del"},
	Short:   "Remove secondary NIC from server",
	Long:    "Remove secondary NIC identified by @net (network ID, CIDR, or IP) from @serverName",
	RunE: func(cmd *cobra.Command, args []string) error {
		var server, net = args[0], args[1]

		if err := client.ServerDelNic(server, net); err != nil {
			exit.Fatalf("failed to remove NIC from %s: %s", server, err)
		}
		return nil
	},
}

// parseNetIP attempts to parse @s as one of
// - CIDR address (address/mask)
// - network address (IP address without mask)
func parseNetIP(s string) (*net.IPNet, error) {
	if ip, network, err := net.ParseCIDR(s); err == nil {
		return network, err
	} else if ip = net.ParseIP(s); ip != nil {
		mask := ip.DefaultMask()
		return &net.IPNet{IP: ip.Mask(mask), Mask: mask}, nil
	}
	return nil, fmt.Errorf("invalid IP/CIDR string %q", s)
}
