package main

/*
 * Print the SBS Backup Policy details for a given server.
 */
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
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <server-ID>\n", path.Base(os.Args[0]))
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

	policies, err := client.SBSgetServerPolicyDetails(flag.Arg(0))
	if err != nil {
		exit.Fatalf("failed to list SBS server policies details of %s: %s", flag.Arg(0), err)
	}

	if len(policies) == 0 {
		fmt.Printf("No server policies found for server %s.\n", flag.Arg(0))
	} else {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoFormatHeaders(false)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetAutoWrapText(false)
		table.SetHeader([]string{"Server", "Server Policy ID", "Status", "Region", "Account",
			"Unsubscribe Date", "Expiration Date"})

		for _, p := range policies {
			table.Append([]string{p.ServerID, p.ID, p.Status, p.StorageRegion, p.ClcAccountAlias,
				fmt.Sprint(p.UnsubscribedDate), fmt.Sprint(p.ExpirationDate)})
		}
		table.Render()
	}
}
