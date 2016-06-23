/*
 * List server templates available at a location
 */
package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
	"github.com/olekukonko/tablewriter"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  Location-Alias\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
		os.Exit(0)
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

	showTemplates(client, flag.Arg(0))
}

func showTemplates(client *clcv2.Client, region string) {
	capa, err := client.GetDeploymentCapabilities(region)
	if err != nil {
		exit.Fatalf("Failed to query deployment capabilities of %s: %s", region, err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoWrapText(false)

	/* Note: not displaying ReservedDrivePaths and DrivePathLength here, I don't understand their use. */
	/* Note: not listing Capabilities here, since the table gets too large for a single screen */
	table.SetHeader([]string{"Template Name", "Description", "OS", "Storage"})

	for _, tpl := range capa.Templates {
		table.Append([]string{tpl.Name, tpl.Description, tpl.OsType, fmt.Sprintf("%d GB", tpl.StorageSizeGB)})
	}
	table.Render()
}
