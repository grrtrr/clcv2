/*
 * Print the list of all regions available in the SBS service.
 */
package main

import (
	"flag"
	"os"
	"strings"

	"github.com/grrtrr/clcv2/clcv2cli"
	"github.com/grrtrr/exit"
	"github.com/olekukonko/tablewriter"
)

func main() {
	flag.Parse()

	client, err := clcv2cli.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	regions, err := client.SBSgetAllRegions()
	if err != nil {
		exit.Fatalf("failed to list SBS regions: %s", err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoWrapText(false)
	table.SetHeader([]string{"Name", "Label", "Messages"})

	for _, re := range regions {
		table.Append([]string{re.Name, re.RegionLabel, strings.Join(re.Messages, ", ")})
	}
	table.Render()
}
