/*
 * List intra-datacenter firewall policies associated with a given account.
 * Optionally filter results to policies associated with a second "destination" account.
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

	fwpl, err := client.GetIntraDataCenterFirewallPolicyList(flag.Arg(0), *dst)
	if err != nil {
		exit.Fatalf("failed to list intra-datacenter firewall policies at %s: %s", flag.Arg(0), err)
	}

	if len(fwpl) == 0 {
		fmt.Printf("Empty result - nothing listed at %s.\n", flag.Arg(0))
	} else if *simple {
		pretty.Println(fwpl)
	} else {
		fmt.Printf("Intra-Datacenter Firewall Policies for %s at %s:\n", client.AccountAlias, strings.ToUpper(flag.Arg(0)))

		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoFormatHeaders(false)
		table.SetAlignment(tablewriter.ALIGN_CENTRE)
		table.SetAutoWrapText(true)

		table.SetHeader([]string{"Source", "Destination", "Ports",
			"Dst Account", "Enabled", "State", "Id",
		})

		for _, p := range fwpl {
			table.Append([]string{
				strings.Join(p.Source, ", "),
				strings.Join(p.Destination, ", "),
				strings.Join(p.Ports, ", "),
				strings.ToUpper(p.DestinationAccount),
				fmt.Sprint(p.Enabled), p.Status, p.ID,
			})
		}
		table.Render()
	}
}
