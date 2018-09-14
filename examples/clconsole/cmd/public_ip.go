/*
 * Add a public IP address to a server
 */
package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	managePublicIps = &cobra.Command{ // Top-level  command
		Use:     "pip",
		Aliases: []string{"public-ip", "ip"},
		Short:   "Manage server Public IPs",
	}

	// List public IP address(es)
	publicIpList = &cobra.Command{
		Use:     "ls  <serverName> [publicIPs...]",
		Aliases: []string{"show", "list"},
		Short:   "List Public IP(s) of a server",
		PreRunE: checkAtLeastArgs(1, "Need a server name"),
		Run: func(cmd *cobra.Command, args []string) {
			var server = args[0] // enforced via PreRunE
			var publicIPs []string

			if len(args) > 1 {
				publicIPs = args[1:]
			}

			if len(publicIPs) == 0 {
				srv, err := client.GetServer(server)
				if err != nil {
					exit.Fatalf("failed to query the public IPs of %s: %s", server, err)
				}

				for _, ip := range srv.Details.IpAddresses {
					if ip.IsPublic() {
						publicIPs = append(publicIPs, ip.Public)
					}
				}

				if len(publicIPs) == 0 {
					fmt.Printf("%s is not associated with any public IP address.\n", server)
					return
				}
			}

			for _, ip := range publicIPs {
				p, err := client.GetPublicIPAddress(server, ip)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to query details of %s public IP %s: %s\n", server, ip, err)
					continue
				}

				fmt.Printf("%s:\n", server)
				fmt.Printf("Public IP:           %s\n", ip)
				fmt.Printf("Internal IP:         %s\n", p.InternalIPAddress)

				if len(p.Ports) > 0 {
					var ports []string

					for _, port := range p.Ports {
						ports = append(ports, port.String())
					}
					fmt.Printf("Port Openings:       %s\n", strings.Join(ports, ", "))
				}
				if len(p.SourceRestrictions) > 0 {
					var srcRes []string

					for _, src := range p.SourceRestrictions {
						srcRes = append(srcRes, src.Cidr)
					}

					fmt.Printf("Source Restrictions: %s\n", strings.Join(srcRes, ", "))
				}
			}
		},
	}

	// Add a new public IP address
	publicIpAdd = &cobra.Command{
		Use:     "add  <serverName>",
		Aliases: []string{"plus"},
		Short:   "Add a public IP to a server",
		PreRunE: checkArgs(1, "Need a server name"),
		Run: func(cmd *cobra.Command, args []string) {
			if len(pipAddFlags.portSp) == 0 { /* default port settings */
				pipAddFlags.portSp.Set("ping")
				pipAddFlags.portSp.Set("ssh")
				pipAddFlags.portSp.Set("http")
			}

			req := clcv2.PublicIPAddress{
				InternalIPAddress:  pipAddFlags.srcIP,
				Ports:              pipAddFlags.portSp,
				SourceRestrictions: pipAddFlags.srcRes,
			}

			if reqID, err := client.AddPublicIPAddress(args[0], &req); err != nil {
				fmt.Fprintf(os.Stderr, "ERROR adding public IP to %s: %s\n", args[0], err)
			} else {
				log.Printf("%s add public IP: %s", args[0], reqID)

				client.PollStatusFn(reqID, intvl, func(s clcv2.QueueStatus) {
					log.Printf("%s add public IP: %s", args[0], s)
				})
			}
		},
	}

	// Flags for adding a Public IP:
	pipAddFlags struct {
		srcIP  string                // Source IP to use for creating the public IP address
		srcRes clcv2.SrcRestrictions // Source CIDR restrictions
		portSp clcv2.PortSpecs       // Port specifications to implement on the new public IP
	}

	// Flags for modifying a Public IP:
	pipModFlags struct {
		keep   bool                  // Whether to keep existing settings
		srcRes clcv2.SrcRestrictions // Source CIDR restrictions
		portSp clcv2.PortSpecs       // Port specifications to implement on the new public IP
	}

	// Modify existing public IP address
	publicIpMod = &cobra.Command{
		Use:     "mod  <serverName>  <public-IP>",
		Aliases: []string{"modify", "update"},
		Short:   "Modify existing server public IP",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := checkArgs(2, "Need a server name and its public IP")(cmd, args); err != nil {
				return err
			} else if len(pipModFlags.portSp) == 0 {
				return errors.Errorf("Need at least 1 port spec (--port argument, can be repeated)")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			if pipModFlags.keep {
				log.Printf("Looking up existing configuration of %s on %s ...", args[1], args[0])
				old, err := client.GetPublicIPAddress(args[0], args[1])
				if err != nil {
					exit.Fatalf("failed to get existing configuration for %s: %s", args[1], err)
				}
				log.Printf("%s existing settings: %v", args[0], old.Ports)
				pipModFlags.portSp = append(pipModFlags.portSp, old.Ports...)
			}

			if reqID, err := client.UpdatePublicIPAddress(args[0], args[1], &clcv2.PublicIPAddress{
				Ports:              pipModFlags.portSp,
				SourceRestrictions: pipModFlags.srcRes,
			}); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to update public IP address %s on %q: %s\n", args[0], args[1], err)
			} else {
				log.Printf("%s modify public IP %s: %s", args[0], args[1], reqID)

				client.PollStatusFn(reqID, intvl, func(s clcv2.QueueStatus) {
					log.Printf("%s modify public IP %s: %s", args[0], args[1], s)
				})
			}
		},
	}

	// Remove public IP
	publicIpDelete = &cobra.Command{
		Use:     "rm <serverName> <publicIP>",
		Aliases: []string{"remove", "delete"},
		Short:   "Remove Public IP from server",
		PreRunE: checkAtLeastArgs(2, "Need a server name and a public IP"),
		Run: func(cmd *cobra.Command, args []string) {
			if reqID, err := client.RemovePublicIPAddress(args[0], args[1]); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to remove public IP address %s from %q: %s\n", args[0], args[1], err)
			} else {
				log.Printf("%s delete public IP %s: %s", args[0], args[1], reqID)

				client.PollStatusFn(reqID, intvl, func(s clcv2.QueueStatus) {
					log.Printf("%s delete public IP %s: %s", args[0], args[1], s)
				})
			}
		},
	}
)

func init() {
	// Add
	publicIpAdd.Flags().Var(&pipAddFlags.srcRes, "restrict", "Restrict source traffic to given CIDR range(s)")
	publicIpAdd.Flags().Var(&pipAddFlags.portSp, "port", "Port spec(s), number(s) or service name(s) (option can be repeated)\n"+
		"        - ping:      use ping or icmp\n"+
		"        - full spec: tcp/20081-20083, udp/554, udp/6080-7000, ...\n"+
		"        - tcp names: rdp, http, https, http-alt, ssh, ftp, ftps, ...\n"+
		"        - tcp ports: 22, 443, 80, 20081-20083, ...\n"+
		"        - DEFAULTS:  ping, ssh, http")
	publicIpAdd.Flags().StringVar(&pipAddFlags.srcIP, "source-ip", "", "Use this existing internal IP on the server")

	// Modify
	publicIpMod.Flags().Var(&pipModFlags.srcRes, "restrict", "Restrict source traffic to given CIDR range(s)")
	publicIpMod.Flags().Var(&pipModFlags.portSp, "port", "Port spec(s), number(s) or service name(s) (option can be repeated)\n"+
		"        - ping:      use ping or icmp\n"+
		"        - full spec: tcp/20081-20083, udp/554, udp/6080-7000, ...\n"+
		"        - tcp names: rdp, http, https, http-alt, ssh, ftp, ftps, ...\n"+
		"        - tcp ports: 22, 443, 80, 20081-20083, ...")
	publicIpMod.Flags().BoolVar(&pipModFlags.keep, "keep", false, "Keep (merge with) existing port settings")

	// Commands
	managePublicIps.AddCommand(publicIpAdd, publicIpList, publicIpMod, publicIpDelete)
	Root.AddCommand(managePublicIps)
}
