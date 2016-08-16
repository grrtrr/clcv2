/*
 * Get the list of data centers that a given account has access to.
 */
package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
	"github.com/olekukonko/tablewriter"
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

	/* Basic data center information */
	ctr, err := client.GetDatacenter(flag.Arg(0), *links)
	if err != nil {
		exit.Fatalf("failed to get datacenter information: %s", err)
	}
	fmt.Printf("%s\n", ctr.Name)

	/* Data centre compute limits */
	compLimits, err := client.GetDatacenterComputeLimits(flag.Arg(0))
	if err != nil {
		exit.Fatalf("failed to get %s compute limits: %s", ctr.Name, err)
	}
	/* Maximum number of networks */
	netLimits, err := client.GetDatacenterNetworkLimits(flag.Arg(0))
	if err != nil {
		exit.Fatalf("failed to get %s network limits: %s", ctr.Name, err)
	}

	fmt.Printf("Limits: CPU: %d, Memory: %s, Storage: %s, Networks: %d\n",
		compLimits.CPU.Value, humanize.Bytes(compLimits.MemoryGB.Value<<30),
		humanize.Bytes(compLimits.StorageGB.Value<<30), netLimits.Networks.Value)

	if *links {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoFormatHeaders(false)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetAutoWrapText(false)

		// See https://www.ctl.io/api-docs/v2/#getting-started-api-v20-links-framework
		table.SetHeader([]string{"Rel", "Href", "Verbs"})
		for _, link := range ctr.Links {
			var verbs string

			if len(link.Verbs) == 0 {
				verbs = "GET"
			} else {
				verbs = strings.Join(link.Verbs, ", ")
			}
			table.Append([]string{link.Rel, link.Href, verbs})
		}
		table.Render()
	}
}
