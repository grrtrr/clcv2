/*
 * Get details of a specific firewall policy associated with a given account in a given data center
 * (an "intra data center firewall policy").
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
	var location = flag.String("l", "", "Data centre location to query")
	var simple   = flag.Bool("simple", false, "Use simple (debugging) output format")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <Location>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 1 || *location == "" {
		flag.Usage()
		os.Exit(1)
	}

	client, err := clcv2.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	fwp, err := client.GetFWPolicy(*location, flag.Arg(0))
	if err != nil {
		exit.Fatalf("Failed to list firewall policy %s at %s: %s", *location, flag.Arg(0), err)
	}

	if *simple {
		pretty.Println(fwp)
	} else {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoFormatHeaders(false)
		table.SetAlignment(tablewriter.ALIGN_CENTRE)
		table.SetAutoWrapText(true)

		table.SetHeader([]string{ "Source", "Destination", "Ports",
			"Dst Account", "Enabled", "State", "Id",
		})
		table.Append([]string{
			strings.Join(fwp.Source, ", "),
			strings.Join(fwp.Destination, ", "),
			strings.Join(fwp.Ports, ", "),
			strings.ToUpper(fwp.DestinationAccount),
			fmt.Sprint(fwp.Enabled), fwp.Status, fwp.Id,
		})
		table.Render()
	}
}
