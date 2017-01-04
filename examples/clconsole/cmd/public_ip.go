/*
 * Add a public IP address to a server
 */
package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/grrtrr/clcv2"
	"github.com/spf13/cobra"
)

var publicIPFlags struct {
	srcIP  string                // source IP to use for creating the public IP address
	srcRes clcv2.SrcRestrictions // source CIDR restrictions
	portSp clcv2.PortSpecs       // port specifications to implement on the new public IP
}

func init() {
	PublicIP.Flags().Var(&publicIPFlags.srcRes, "restrict", "Restrict source traffic to given CIDR range(s)")
	PublicIP.Flags().Var(&publicIPFlags.portSp, "port", "Port spec(s), number(s) or service name(s) (option can be repeated)\n"+
		"        - ping:      use ping or icmp\n"+
		"        - full spec: tcp/20081-20083, udp/554, udp/6080-7000, ...\n"+
		"        - tcp names: rdp, http, https, http-alt, ssh, ftp, ftps, ...\n"+
		"        - tcp ports: 22, 443, 80, 20081-20083, ...\n"+
		"        - DEFAULTS:  ping, ssh, http")
	PublicIP.Flags().StringVar(&publicIPFlags.srcIP, "source-ip", "", "Use this existing internal IP on the server")

	Root.AddCommand(PublicIP)
}

var PublicIP = &cobra.Command{
	Use:     "public <serverName>",
	Aliases: []string{"public_ip", "pip"},
	Short:   "Add a public IP to a server",
	PreRunE: checkArgs(1, "Need a server name"),
	Run: func(cmd *cobra.Command, args []string) {
		if len(publicIPFlags.portSp) == 0 { /* default port settings */
			publicIPFlags.portSp.Set("ping")
			publicIPFlags.portSp.Set("ssh")
			publicIPFlags.portSp.Set("http")
		}

		req := clcv2.PublicIPAddress{
			InternalIPAddress:  publicIPFlags.srcIP,
			Ports:              publicIPFlags.portSp,
			SourceRestrictions: publicIPFlags.srcRes,
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
