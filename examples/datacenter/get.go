/*
 * Get the list of data centers that a given account has access to.
 */
package main

import (
	"github.com/olekukonko/tablewriter"
	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
	"strings"
	"path"
	"flag"
	"fmt"
	"os"
)

func main() {
	var links = flag.Bool("l", true, "List the group links as well")

	flag.Usage = func() {
                fmt.Fprintf(os.Stderr, "usage: %s [options]  <Location>\n", path.Base(os.Args[0]))
                flag.PrintDefaults()
        }

        flag.Parse()
        if flag.NArg() != 1 {
                flag.Usage()
                os.Exit(1)
        }

	client, err := clcv2.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	ctr, err := client.GetDatacenter(flag.Arg(0), *links)
	if err != nil {
		exit.Fatalf("Failed to get datacenter information: %s", err)
	}

	fmt.Printf("%s\n", ctr.Name)
	if *links {
			table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoFormatHeaders(false)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetAutoWrapText(false)

		// See https://www.ctl.io/api-docs/v2/#getting-started-api-v20-links-framework
		table.SetHeader([]string{ "Rel", "Href", "Verbs" } )
		for _, link := range ctr.Links {
			var verbs string

			if len(link.Verbs) == 0 {
				verbs = "GET"
			} else {
				verbs = strings.Join(link.Verbs, ", ")
			}
			table.Append([]string{ link.Rel, link.Href, verbs })
		}
		table.Render()
	}
}

