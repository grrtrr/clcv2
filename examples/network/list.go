/*
 * List networks available for a given account in a given data center.
 */
package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/grrtrr/clcv2/clcv2cli"
	"github.com/grrtrr/exit"
	"github.com/kr/pretty"
	"github.com/olekukonko/tablewriter"
)

func main() {
	var simple = flag.Bool("simple", false, "Use simple (debugging) output format")

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

	networks, err := client.GetNetworks(flag.Arg(0))
	if err != nil {
		exit.Fatalf("failed to list networks: %s", err)
	}

	if len(networks) == 0 {
		println("Empty result.")
	} else if *simple {
		pretty.Println(networks)
	} else {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoFormatHeaders(false)
		table.SetAlignment(tablewriter.ALIGN_RIGHT)
		table.SetAutoWrapText(false)

		table.SetHeader([]string{"CIDR", "Gateway", "VLAN", "Name", "Description", "Type", "ID"})
		for _, l := range networks {
			table.Append([]string{l.Cidr, l.Gateway, fmt.Sprint(l.Vlan), l.Name, l.Description, l.Type, l.Id})
		}
		table.Render()
	}
}
