/*
 * List servers with a given data centre.
 * NOTE: the data centre name is the one returned by the SBS 'List Data Centers' command, NOT
 *       the one usually used for CLC data centres! However, when I tested this, it seems that
 *       the standard CLC data centre identifiers (e.g. "UC1", or "ny1") also work.
 */
package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
	"github.com/olekukonko/tablewriter"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  '<data centre name>'\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(0)
	}

	client, err := clcv2.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	servers, err := client.SBSgetServersByDatacenter(flag.Arg(0))
	if err != nil {
		exit.Fatalf("failed to list SBS servers in %q: %s", flag.Arg(0), err)
	}

	if len(servers) == 0 {
		fmt.Printf("No servers found in %s.\n", flag.Arg(0))
	} else {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoFormatHeaders(false)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetAutoWrapText(false)
		table.SetHeader([]string{"Server Name"})

		// The query returns servers in lower-case format. Upper-case them for consistency with CLC naming.
		for _, server := range servers {
			table.Append([]string{strings.ToUpper(server)})
		}
		table.Render()
	}
}
