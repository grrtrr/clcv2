/*
 * Lists all hardware groups associated with a given location (and account alias).
 */
package main

import (
	"github.com/olekukonko/tablewriter"
	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
	"github.com/kr/pretty"
	"path"
	"flag"
	"fmt"
	"os"
)

func main() {
	var simple = flag.Bool("simple", false, "Use simple (debugging) output format")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <Location>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()

	/* The Location argument is always required */
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	client, err := clcv2.NewClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	rootNode, err := client.GetGroups(flag.Arg(0))
	if err != nil {
		exit.Fatalf("Failed to list hardware groups: %s", err)
	}

	if *simple {
		pretty.Println(rootNode)
	} else {
		fmt.Printf("%s in %s (%s, %d servers), ID %s:\n", rootNode.Name, rootNode.LocationId,
			   rootNode.Status, rootNode.Serverscount, rootNode.Id)

		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoFormatHeaders(false)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetAutoWrapText(true)

		table.SetHeader([]string{ "Name", "UUID", "Description", "#Servers", "Type"})
		for _, g := range rootNode.Groups {
			table.Append([]string{ g.Name, g.Id, g.Description, fmt.Sprint(g.Serverscount), g.Type })
		}
		table.Render()
	}
}
