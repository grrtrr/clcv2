package cmd

import (
	"log"

	"github.com/grrtrr/exit"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	var (
		nic = &cobra.Command{
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

		addNICFlags struct {
			ip string // IP address to assign to the secondary NIC
		}

		addNIC = &cobra.Command{
			Use:   "add  <serverName>  <net (ID | CIDR | IP)>",
			Short: "Add a secondary NIC to server",
			Long:  "Add a secondary NIC to @server on network @net (using network ID, CIDR, or IP)",
			Run: func(cmd *cobra.Command, args []string) {
				netID, err := resolveNet(args[1], location)
				if err != nil {
					exit.Errorf("failed to resolve %s: %s", args[1], err)
				}
				log.Printf("Adding %s NIC on network %s ...")
				if err = client.ServerAddNic(args[0], netID, addNICFlags.ip); err != nil {
					exit.Fatalf("failed to add NIC to %s: %s", args[0], err)
				}
			},
		}

		removeNIC = &cobra.Command{
			Use:     "remove  <serverName>  <net (ID | CIDR | IP)>",
			Aliases: []string{"rm", "del"},
			Short:   "Remove secondary NIC from server",
			Long:    "Remove secondary NIC identified by @net (network ID, CIDR, or IP) from @serverName",
			Run: func(cmd *cobra.Command, args []string) {
				netID, err := resolveNet(args[1], location)
				if err != nil {
					exit.Errorf("failed to resolve %s: %s", args[1], err)
				}
				log.Printf("Deleting %s NIC on network %s ...", args[0], netID)
				if err = client.ServerDelNic(args[0], netID); err != nil {
					exit.Fatalf("failed to remove NIC from %s: %s", args[0], err)
				}
			},
		}
	)

	addNIC.Flags().StringVar(&addNICFlags.ip, "ip", "", "IP address to use with NIC (optional, default is automatic assignment)")
	nic.AddCommand(addNIC)
	nic.AddCommand(removeNIC)

	Root.AddCommand(nic)
}
