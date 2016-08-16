/*
 * For a given Hardware Group name at a given Location, print the corresponding HW Group UUID.
 */
package main

import (
	"github.com/olekukonko/tablewriter"
	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
	"path"
	"flag"
	"fmt"
	"os"
)

func main() {
	var location = flag.String("l", "", "Data center location of @Group-Name")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <Group-Name>\n", path.Base(os.Args[0]))
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

	g, err := client.GetGroupByName(flag.Arg(0), *location)
	if err != nil {
		exit.Fatalf("failed to look up %s: %s", flag.Arg(0), err)
	} else if g == nil {
		exit.Fatalf("No such group %s", flag.Arg(0))
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoWrapText(true)
	table.SetHeader([]string{ "Name", "UUID", "Description", "#Servers", "Type"})

	table.Append([]string{ g.Name, g.Id, g.Description, fmt.Sprint(g.Serverscount), g.Type })

	table.Render()
}
