/*
 * Prints out networking information for a given server.
 */
package main

import (
	"github.com/olekukonko/tablewriter"
	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
_	"strings"
	"path"
	"flag"
	"fmt"
	"os"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <Server Name>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	client, err := clcv2.NewClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	server, err := client.GetServer(flag.Arg(0))
	if err != nil {
		exit.Fatalf("Failed to list details of server %q: %s", flag.Arg(0), err)
	}

	var public_ips, private_ips []string

	// Get networks for the location the server resides in
	networks, err := client.GetNetworks(server.LocationId)
	if err != nil {
		exit.Fatalf("Failed to list networks: %s", err)
	}

	for _, ip := range server.Details.IpAddresses {
		if ip.Public != "" {
			public_ips = append(public_ips, ip.Public)
		}
		if ip.Internal != "" {
			private_ips = append(private_ips, ip.Internal)
		}
	}
	/*
		if len(public_ips) > 0 {
			fmt.Printf("Public IPs: none\n")
		} else {
			fmt.Printf("Public IPs: %s\n", strings.Join(public_ips, ", "))
		}
		fmt.Printf("Private IPs: %s\n", strings.Join(private_ips, ", "))
*/
		
	for _, ip_string := range private_ips {
		n, err := clcv2.NetworkByIP(ip_string, networks)
		if err != nil {
			exit.Fatalf("Failed to identify network for %s: %s", ip_string, err)
		}
		
		fmt.Printf("Network for %s of %s in %s:\n", ip_string, server.Name, server.LocationId)
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoFormatHeaders(false)
		table.SetAlignment(tablewriter.ALIGN_RIGHT)
		table.SetAutoWrapText(true)
			
		table.SetHeader([]string{
			"Name", "Description", "Type", "ID",
			"CIDR", "Netmask", "Gateway", "VLAN",
		})
		table.Append([]string{ n.Name, n.Description, n.Type,  n.Id,
			n.Cidr, n.Netmask, n.Gateway, fmt.Sprint(n.Vlan) })
		table.Render()
	}
}
