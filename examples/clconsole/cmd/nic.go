package cmd

import (
	"log"

	"github.com/grrtrr/exit"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	var (
		nic = &cobra.Command{ // Top-level NIC command
			Use:   "nic",
			Short: "Manage server NICs",
			Long:  "Add or remove server secondary network interface",
			PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
				// Both commands take <serverName> <net ID | name | CIDR>
				if len(args) < 2 {
					return errors.Errorf("Need a server name and a network specifier " +
						"(network ID/name/CIDR, or IP on the network) for the secondary NIC")
				}
				setLocationBasedOnServerName(args[0])
				return nil
			},
		}

		addNICFlags struct {
			ip string // IP address to assign to the secondary NIC
		}

		addNIC = &cobra.Command{
			Use:   "add  <serverName>  <net (ID | Name | CIDR | IP)>",
			Short: "Add a secondary NIC to server",
			Long:  "Add a secondary NIC to @server on network @net (using network ID, name, CIDR, or and IP on the network)",
			Run: func(cmd *cobra.Command, args []string) {
				var server, netID = args[0], args[1]

				network, err := resolveNet(netID, conf.Location)
				if err != nil {
					exit.Errorf("failed to resolve %s: %s", netID, err)
				} else if network != nil {
					netID = network.Id
				}

				log.Printf("Adding %s NIC on network %s ...", server, netID)
				if err = client.ServerAddNic(server, netID, addNICFlags.ip); err != nil {
					log.Fatalf("failed to add NIC to %s: %s", server, err)
				}
				log.Printf("Successfully added NIC to server %s", server)
			},
		}

		removeNIC = &cobra.Command{
			Use:     "remove  <serverName>  <net (ID | Name | CIDR | IP)>",
			Aliases: []string{"rm", "del"},
			Short:   "Remove secondary NIC from server",
			Long:    "Remove secondary NIC identified by @net (network ID, name, CIDR, or an IP on the network) from @serverName",
			Run: func(cmd *cobra.Command, args []string) {
				var server, netID = args[0], args[1]

				network, err := resolveNet(netID, conf.Location)
				if err != nil {
					exit.Errorf("failed to resolve %s: %s", netID, err)
				} else if network != nil {
					netID = network.Id
				}

				log.Printf("Deleting %s NIC on network %s ...", server, netID)
				if err = client.ServerDelNic(server, netID); err != nil {
					log.Fatalf("failed to remove NIC from %s: %s", server, err)
				}
				log.Printf("Successfully removed NIC from server %s", server)
			},
		}
	)

	addNIC.Flags().StringVar(&addNICFlags.ip, "ip", "", "IP address to use with NIC (optional, default is automatic assignment)")
	nic.AddCommand(addNIC)
	nic.AddCommand(removeNIC)

	Root.AddCommand(nic)
}
