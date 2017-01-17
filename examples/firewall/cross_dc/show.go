/*
 * Display details of a single cross-datacenter firewall policy.
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
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <Location> <PolicyID>\n", path.Base(os.Args[0]))
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

	p, err := client.GetCrossDataCenterFirewallPolicy(flag.Arg(0), flag.Arg(1))
	if err != nil {
		exit.Fatalf("failed to list %s firewall policy %s: %s", flag.Arg(0), flag.Arg(1), err)
	}

	fmt.Printf("Details of %s cross-datacenter Firewall Policy %s:\n", strings.ToUpper(flag.Arg(0)), flag.Arg(1))

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetAlignment(tablewriter.ALIGN_CENTER)
	table.SetAutoWrapText(false)

	table.SetHeader([]string{"Data Center", "Network", "Account", "Status", "ID"})

	enabledStr := "*"
	if !p.Enabled {
		enabledStr = "-"
	}
	table.Append([]string{
		strings.ToUpper(fmt.Sprintf("%s => %s", p.SourceLocation, p.DestLocation)),
		fmt.Sprintf("%18s => %-18s", p.SourceCIDR, p.DestCIDR),
		strings.ToUpper(fmt.Sprintf("%-4s => %4s", p.SourceAccount, p.DestAccount)),
		fmt.Sprintf("%s%s", enabledStr, p.Status),
		p.ID,
	})

	table.Render()
}
