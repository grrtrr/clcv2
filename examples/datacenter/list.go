/*
 * Get the list of data centers that a given account has access to.
 */
package main

import (
	"github.com/olekukonko/tablewriter"
	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
	"strings"
	"flag"
	"fmt"
	"os"
)

func main() {
	var links = flag.Bool("l", false, "List the Link References as well")

	flag.Parse()

	client, err := clcv2.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	datacenters, err := client.GetLocations()
	if err != nil {
		exit.Fatalf("failed to get datacenter information: %s", err)
	}

	if *links {
		for _, ctr := range datacenters {
			fmt.Println("\n", ctr.Name)

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
	} else {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoFormatHeaders(false)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetAutoWrapText(false)
		table.SetHeader([]string{ "Id", "Name", } )

		for _, ctr := range datacenters {
			table.Append([]string{ ctr.Id, ctr.Name })
		}
		table.Render()
	}
}

