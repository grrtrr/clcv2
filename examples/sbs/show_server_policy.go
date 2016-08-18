/*
 * Print details of a single Server Policy given its server-policy-ID.
 */
package main

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
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <Server-Policy-ID>\n", path.Base(os.Args[0]))
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

	p, err := client.SBSgetServerPolicy(flag.Arg(0))
	if err != nil {
		exit.Fatalf("failed to list SBS server policy %s: %s", flag.Arg(0), err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoWrapText(false)
	table.SetHeader([]string{"Server", "Server Policy ID", "Account Policy ID",
		"Status", "Region", "Account", "Unsubscribe Date", "Expiration Date"})

	table.Append([]string{p.ServerID, p.ID, p.AccountPolicyID,
		p.Status, p.StorageRegion, p.ClcAccountAlias,
		fmt.Sprint(p.UnsubscribedDate), fmt.Sprint(p.ExpirationDate)})
	table.Render()
}
