/*
 * Prints out networking information for a given server.
 */
package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/clcv2/clcv2cli"
	"github.com/grrtrr/exit"
	"github.com/olekukonko/tablewriter"
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

	client, err := clcv2cli.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	server, err := client.GetServer(flag.Arg(0))
	if err != nil {
		exit.Fatalf("failed to list details of server %q: %s", flag.Arg(0), err)
	}

	// Get networks for the location the server resides in
	networks, err := client.GetNetworks(server.LocationId)
	if err != nil {
		exit.Fatalf("failed to list networks: %s", err)
	}

	var private_ips []string
	var public_ips []clcv2.ServerIPAddress

	for _, ip := range server.Details.IpAddresses {
		if ip.Public != "" {
			public_ips = append(public_ips, ip)
		}
		if ip.Internal != "" {
			private_ips = append(private_ips, ip.Internal)
		}
	}

	fmt.Printf("IP addresses of %s:\n", server.Name)

	if len(public_ips) > 0 {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoFormatHeaders(false)
		table.SetAlignment(tablewriter.ALIGN_CENTRE)
		table.SetAutoWrapText(true)
		table.SetHeader([]string{"Public IP", "Via"})
		for _, ip := range public_ips {
			table.Append([]string{ip.Public, ip.Internal})
		}
		table.Render()
		fmt.Println()
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetAlignment(tablewriter.ALIGN_CENTRE)
	table.SetAutoWrapText(true)

	table.SetHeader([]string{fmt.Sprintf("%s IP", server.Name),
		"CIDR", "Gateway", "VLAN", "Type", "Description",
		"Network Name", "Network ID",
	})

	for _, ip_string := range private_ips {
		n, err := clcv2.NetworkByIP(ip_string, networks)
		if err != nil {
			exit.Fatalf("failed to identify network for %s: %s", ip_string, err)
		} else if n == nil {
			exit.Fatalf("no matching network found for %s in %s", ip_string, server.LocationId)
		}
		table.Append([]string{
			ip_string, n.Cidr, n.Gateway, fmt.Sprint(n.Vlan),
			n.Type, n.Description, n.Name, n.Id,
		})
	}
	table.Render()
}
