package cmd

import (
	"fmt"

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
