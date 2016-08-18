package main

/*
 * Print the list of SBS server policies associated with a given Account Policy ID.
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
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <Account-Policy-ID>\n", path.Base(os.Args[0]))
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

	policies, err := client.SBSgetServerPolicies(flag.Arg(0))
	if err != nil {
		exit.Fatalf("failed to list SBS server policies of Account Policy ID %s: %s", flag.Arg(0), err)
	}

	if len(policies) == 0 {
		fmt.Printf("No server policies found for Account Policy %s.\n", flag.Arg(0))
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
