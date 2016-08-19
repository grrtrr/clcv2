/*
 * Print details of a single Server Policy given its server-policy-ID, or server name.
 */
package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/clcv2/utils"
	"github.com/grrtrr/exit"
	"github.com/olekukonko/tablewriter"
)

func main() {
	var policies []clcv2.SBSServerPolicy

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <Server-Policy-ID | Server-Name>\n", path.Base(os.Args[0]))
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

	if utils.LooksLikeServerName(flag.Arg(0)) {
		// Note: when using the server name instead of the Server Policy ID, sometimes the Status field is empty.
		policies, err = client.SBSgetServerPolicyDetails(flag.Arg(0))
		if err != nil {
			exit.Fatalf("failed to list SBS policies for server %s: %s", flag.Arg(0), err)
		}
	} else {
		// Query by Server Policy ID
		p, err := client.SBSgetServerPolicy(flag.Arg(0))
		if err != nil {
			exit.Fatalf("failed to list SBS Server Policy %s: %s", flag.Arg(0), err)
		}
		policies = []clcv2.SBSServerPolicy{*p}
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoWrapText(false)
	table.SetHeader([]string{"Server", "Server Policy ID", "Account Policy ID",
		"Status", "Region", "Account", "Unsubscribe Date", "Expiration Date"})

	for _, p := range policies {
		table.Append([]string{p.ServerID, p.ID, p.AccountPolicyID,
			p.Status, p.StorageRegion, p.ClcAccountAlias,
			fmt.Sprint(p.UnsubscribedDate), fmt.Sprint(p.ExpirationDate)})
	}
	table.Render()
}
