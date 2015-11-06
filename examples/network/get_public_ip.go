/*
 * Print details of a public IP address on a server.
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
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <server-name> <public-IP>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(1)
	}

	client, err := clcv2.NewClient()
	if err != nil {
		exit.Fatal(err.Error())
	}


	p, err := client.GetPublicIPAddress(flag.Arg(0), flag.Arg(1))
	if err != nil {
		exit.Fatalf("Failed to query public IP address to %q: %s", flag.Arg(0), err)
	}

	fmt.Printf("Details of public IP %s on %s:\n", flag.Arg(1), flag.Arg(0))
	fmt.Printf("Associated internal IP: %s\n", p.InternalIPAddress)

	if len(p.Ports) > 0 {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoFormatHeaders(false)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetAutoWrapText(false)

		table.SetHeader([]string{ "Proto", "Port" })
		for _, port := range p.Ports  {
			var portSpec = fmt.Sprint(port.Port)

			if port.PortTo != 0 {
				portSpec += fmt.Sprintf("-%d", port.PortTo)
			}

			table.Append([]string{ port.Protocol, portSpec })
		}
		table.Render()
	}
	if len(p.SourceRestrictions) > 0 {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoFormatHeaders(false)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetAutoWrapText(false)

		table.SetHeader([]string{ "Source Traffic Restriction" })
		for _, src := range p.SourceRestrictions  {
			table.Append([]string{ src.Cidr })
		}
		table.Render()
	}
}
