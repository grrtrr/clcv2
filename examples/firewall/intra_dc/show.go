/*
 * Get details of a specific intra-datacenter firewall policy associated with a given account.
 */
package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/grrtrr/clcv2/clcv2cli"
	"github.com/grrtrr/exit"
	"github.com/olekukonko/tablewriter"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <Location> <Policy-ID>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(1)
	}

	client, err := clcv2cli.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	fwp, err := client.GetIntraDataCenterFirewallPolicy(flag.Arg(0), flag.Arg(1))
	if err != nil {
		exit.Fatalf("failed to list intra-datacenter firewall policy %s at %s: %s", flag.Arg(0), flag.Arg(1), err)
	}

	fmt.Printf("Details of intra-datacenter Firewall Policy %s at %s:\n", flag.Arg(1), strings.ToUpper(flag.Arg(0)))
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetAlignment(tablewriter.ALIGN_CENTRE)
	table.SetAutoWrapText(true)

	table.SetHeader([]string{"Source", "Destination", "Ports",
		"Dst Account", "Enabled", "State", "Id",
	})
	table.Append([]string{
		strings.Join(fwp.Source, ", "),
		strings.Join(fwp.Destination, ", "),
		strings.Join(fwp.Ports, ", "),
		strings.ToUpper(fwp.DestinationAccount),
		fmt.Sprint(fwp.Enabled), fwp.Status, fwp.ID,
	})
	table.Render()
}
