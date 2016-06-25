/*
 * Get the details of a specific network in a given data center for a given account.
 */
package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
	"github.com/kr/pretty"
	"github.com/olekukonko/tablewriter"
)

func main() {
	var query = flag.String("q", "none", "Filter IP addresses; one of 'none', 'claimed', 'free', or 'all'")
	var location = flag.String("l", "", "Data centre alias of the network")
	var simple = flag.Bool("simple", false, "Use simple (debugging) output format")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <Network-ID (hex)>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	/* Location is required (despite hex id), an empty location leads to a "404 Not Found" response. */
	if flag.NArg() != 1 || *location == "" {
		flag.Usage()
		os.Exit(1)
	} else if !inStringArray(*query, "none", "claimed", "free", "all") {
		exit.Errorf("Invalid IP query %q. Try -h")
	}

	client, err := clcv2.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	details, err := client.GetNetworkDetails(*location, flag.Arg(0), *query)
	if err != nil {
		exit.Fatalf("Failed to query network details of %s: %s", flag.Arg(0), err)
	}

	if *simple {
		pretty.Println(details)
	} else {
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
}

/* go replacement for python 'x in list' */
func inStringArray(s string, list ...string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
