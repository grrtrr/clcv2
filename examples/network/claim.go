/*
 * Claim a new network in a given data centre.
 */
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/clcv2/clcv2cli"
	"github.com/olekukonko/tablewriter"
)

func main() {
	var location string

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <data-centre>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}
	location = strings.ToUpper(flag.Arg(0))

	client, err := clcv2cli.NewCLIClient()
	if err != nil {
		log.Fatal(err.Error())
	}

	id, err := client.ClaimNetwork(location, func(s clcv2.QueueStatus) {
		log.Printf("claiming new network in %s: %s", location, s)
	})
	if err != nil {
		log.Fatalf("failed to claim a new network in %q: %s", location, err)
	}
	log.Printf("Successfully claimed new network %s in %s", id, location)

	details, err := client.GetNetworkDetails(location, id, "all")
	if err != nil {
		log.Fatalf("failed to query network details of network %s: %s", id, err)
	}
	printNetworkDetails(details)
}

// printNetworkDetails pretty-prints @details
// FIXME: duplicated from details.go
func printNetworkDetails(details clcv2.NetworkDetails) {
	fmt.Printf("Details of %s (%s):\n", details.Name, details.Description)
	fmt.Printf("CIDR:    %s\n", details.Cidr)
	fmt.Printf("Gateway: %s\n", details.Gateway)
	fmt.Printf("Type:    %s\n", details.Type)
	fmt.Printf("VLAN:    %d\n", details.Vlan)

	if len(details.IpAddresses) > 0 {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoFormatHeaders(false)
		table.SetAlignment(tablewriter.ALIGN_RIGHT)
		table.SetAutoWrapText(false)

		table.SetHeader([]string{"Address", "Claimed", "Server", "Type"})
		for _, i := range details.IpAddresses {
			table.Append([]string{i.Address, fmt.Sprint(i.Claimed), i.Server, i.Type})
		}
		table.Render()
	}
}
