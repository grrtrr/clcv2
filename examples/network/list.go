/*
 * List networks available for a given account in a given data center.
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
	"github.com/grrtrr/exit"
	"github.com/olekukonko/tablewriter"
)

func main() {
	var location string

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <Location>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	client, err := clcv2cli.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	if flag.NArg() != 1 {
		location = client.LocationAlias
	} else {
		location = flag.Arg(0)
	}

	fmt.Printf("Networks visible to %s account in %s:\n", client.AccountAlias, strings.ToUpper(location))
	showNetworks(client, location, client.AccountAlias)
	if client.AccountAlias != client.RegisteredAccountAlias() {
		fmt.Printf("Networks visible to parent %s account:\n", client.RegisteredAccountAlias())
		showNetworks(client, location, client.RegisteredAccountAlias())
	}
}

// showNetworks shows networks visible to @account in data centre location @location.
func showNetworks(client *clcv2.CLIClient, location, account string) {
	networks, err := client.GetNetworks(location, account)
	if err != nil {
		exit.Fatalf("failed to list networks in %s: %s", location, err)
	}

	if len(networks) == 0 {
		println("Empty result.")
	} else {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoFormatHeaders(false)
		table.SetAlignment(tablewriter.ALIGN_RIGHT)
		table.SetAutoWrapText(false)

		table.SetHeader([]string{"CIDR", "Gateway", "VLAN", "Name", "Description", "Type", "ID"})
		for _, l := range networks {
			table.Append([]string{l.Cidr, l.Gateway, fmt.Sprint(l.Vlan), l.Name, l.Description, l.Type, l.Id})
		}
		table.Render()
	}
}
