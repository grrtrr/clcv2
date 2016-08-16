/*
 * Print the list of CLC data centres.
 */
package main

import (
	"flag"
	"os"

	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
	"github.com/olekukonko/tablewriter"
)

func main() {
	flag.Parse()

	client, err := clcv2.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	centres, err := client.SBSgetDatacenters()
	if err != nil {
		exit.Fatalf("failed to list SBS data centres: %s", err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoWrapText(false)
	table.SetHeader([]string{"Data Centre"})

	for _, ctr := range centres {
		table.Append([]string{ctr})
	}
	table.Render()
}
