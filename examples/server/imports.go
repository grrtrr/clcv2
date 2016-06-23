/*
 * Gets the list of available servers that can be imported.
 * Use this API operation when you want to get the list of available OVFs that can be imported
 * with the Import Server API. These OVFs are ones that have been uploaded to your FTP server
 * as described in Using Self-Service VM Import:
 * http://www.ctl.io/knowledge-base/servers/using-self-service-vm-import/
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
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <Location ID>\n", path.Base(os.Args[0]))
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

	imports, err := client.GetServerImports(flag.Arg(0))
	if err != nil {
		exit.Fatalf("Failed to list server imports available at %q: %s", flag.Arg(0), err)
	}

	if len(imports) == 0 {
		fmt.Printf("No imports listed at %s.\n", flag.Arg(0))
	} else {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoFormatHeaders(false)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetAutoWrapText(true)

		table.SetHeader([]string{ "Id", "Name", "Storage/GB", "#CPU", "Membory/MB" })

		for _, i := range imports {
			table.Append([]string{
				i.Id, i.Name, fmt.Sprint(i.StorageSizeGb),
				fmt.Sprint(i.CpuCount), fmt.Sprint(i.MemorySizeMb),
			})
		}
		table.Render()
	}
}
