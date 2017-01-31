package cmd

import (
	"github.com/grrtrr/exit"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	var nic = &cobra.Command{
		Use:   "nic",
		Short: "Manage server NICs",
		Long:  "Add or remove server secondary network interface",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return errors.Errorf("Need one or more sources (group/server) and a destination folder")
			}
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
	Use:     "add  <serverName>  <net>",
	Short:   "Add a secondary NIC to server",
	Long:    "Add a secondary NIC to @server on network @net (using network name or ID)",
	PreRunE: checkArgs(2, "Need a server name and a network ID or name for the secondary NIC"),
	RunE: func(cmd *cobra.Command, args []string) error {
		var server, net = args[0], args[1]

		if err := client.ServerAddNic(server, net, addNICFlags.ip); err != nil {
			exit.Fatalf("failed to add NIC to %s: %s", server, err)
		}
		return nil
	},
}

var removeNIC = &cobra.Command{
	Use:     "remove  <serverName>  <net>",
	Aliases: []string{"rm", "del"},
	Short:   "Remove secondary NIC from server",
	Long:    "Remove secondary NIC identified by @net (network ID or name) from @serverName",
	PreRunE: checkArgs(2, "Need a server name and a network ID or name for the secondary NIC"),
	RunE: func(cmd *cobra.Command, args []string) error {
		var server, net = args[0], args[1]

		if err := client.ServerDelNic(server, net); err != nil {
			exit.Fatalf("failed to remove NIC from %s: %s", server, err)
		}
		return nil
	},
}
