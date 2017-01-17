/*
 * List cross-datacenter firewall policies.
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
	"github.com/kr/pretty"
	"github.com/olekukonko/tablewriter"
)

func main() {
	var simple = flag.Bool("simple", false, "Use simple (debugging) output format")
	var dst = flag.String("dst", "", "Destination account to filter policies by")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <Location>\n", path.Base(os.Args[0]))
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

	fwpl, err := client.GetCrossDataCenterFirewallPolicyList(flag.Arg(0), *dst)
	if err != nil {
		exit.Fatalf("failed to list firewall policies at %s: %s", flag.Arg(0), err)
	}

	if len(fwpl) == 0 {
		fmt.Printf("Empty result - nothing listed at %s.\n", flag.Arg(0))
	} else if *simple {
		pretty.Println(fwpl)
	} else {
		fmt.Printf("Cross-Datacenter Firewall Policies at %s:\n", strings.ToUpper(flag.Arg(0)))

		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoFormatHeaders(false)
		table.SetAlignment(tablewriter.ALIGN_CENTER)
		table.SetAutoWrapText(false)

		table.SetHeader([]string{"Data Center", "Network", "Account", "Status", "ID"})
		for _, p := range fwpl {
			var enabledStr = "*"

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
		}
		table.Render()
	}
}
