/*
 * List firewall policies associated with a given account in a given data center
 * ("intra data center firewall policies").
 * Optionally filter results to policies associated with a second "destination" account.
 */
package main

import (
	"github.com/olekukonko/tablewriter"
	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
	"github.com/kr/pretty"
	"strings"
	"path"
	"flag"
	"fmt"
	"os"
)

func main() {
	var simple = flag.Bool("simple", false, "Use simple (debugging) output format")
	var dst    = flag.String("dst", "",     "Destination account to filter policies by")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <Location>\n", path.Base(os.Args[0]))
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

	
	fwpl, err := client.GetFWPolicyList(flag.Arg(0), *dst)
	if err != nil {
		exit.Fatalf("Failed to list firewall policies at %s: %s", flag.Arg(0), err)
	}

	if len(fwpl) == 0 {
		fmt.Printf("Empty result - nothing listed at %s.\n", flag.Arg(0))
	} else if *simple {
		pretty.Println(fwpl)
	} else {
		fmt.Printf("Intra Data Center Firewall Policies at %s:\n", flag.Arg(0))

		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoFormatHeaders(false)
		table.SetAlignment(tablewriter.ALIGN_CENTRE)
		table.SetAutoWrapText(true)

		table.SetHeader([]string{ "State", "Source", "Destination",
					  "Ports", "Dst Account", "Id",
		})
		for _, p := range fwpl {
			var stateStr = p.Status

			if p.Enabled {
				stateStr += "*"
			}
			table.Append([]string{ stateStr,
				strings.Join(p.Source, ", "),
				strings.Join(p.Destination, ", "),
				strings.Join(p.Ports, ", "),
				strings.ToUpper(p.DestinationAccount), p.Id,
			})
		}

		table.Render()
	}
}
