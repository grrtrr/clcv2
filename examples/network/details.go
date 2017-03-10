/*
 * Get the details of a specific network in a given data center for a given account.
 */
package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/clcv2/clcv2cli"
	"github.com/grrtrr/clcv2/utils"
	"github.com/grrtrr/exit"
	"github.com/olekukonko/tablewriter"
)

func main() {
	var location = flag.String("l", os.Getenv("CLC_LOCATION"), "Data centre alias (needed to resolve IDs)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options] -l <Location>  <Network-ID (hex)>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()

	/* Location is required (despite hex id), an empty location leads to a "404 Not Found" response. */
	if flag.NArg() != 1 || *location == "" {
		flag.Usage()
		os.Exit(1)
	}

	client, err := clcv2cli.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	// Always query all IPs. It does not increase the overhead much and is the most comprehensive variant.
	// Other possible query types are: "claimed" , "free", and "none" (does not list any IP addresses).
	details, err := client.GetNetworkDetails(*location, flag.Arg(0), "all")
	if err != nil {
		exit.Fatalf("failed to query network details of %s: %s", flag.Arg(0), err)
	}
	printNetworkDetails(details)
}

// printNetworkDetails pretty-prints @details
func printNetworkDetails(details clcv2.NetworkDetails) {
	var table = tablewriter.NewWriter(os.Stdout)
	var claimed []clcv2.IpAddressDetails
	var free []string

	if len(details.IpAddresses) > 0 {
		for _, addr := range details.IpAddresses {
			if addr.Claimed {
				claimed = append(claimed, addr)
			} else {
				free = append(free, addr.Address)
			}
		}
	}

	fmt.Printf("Details of network %q", details.Name)
	if details.Description != details.Name {
		fmt.Printf(" (%s)", details.Description)
	}
	fmt.Printf(", ID %s:\n", details.Id)
	table.SetHeader([]string{"CIDR", "Gateway", fmt.Sprintf("Free IPs (%d)", len(free)), "Type", "VLAN"})
	table.Append([]string{
		details.Cidr,
		details.Gateway,
		strings.Join(utils.CollapseIpRanges(free), ", "),
		details.Type,
		fmt.Sprint(details.Vlan),
	})
	table.Render()

	if len(claimed) > 0 {
		table = tablewriter.NewWriter(os.Stdout)
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
