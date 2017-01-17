/*
 * Print the amount of stored data used by a single server, identified by its name or its Server Policy ID.
 */
package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/clcv2/clcv2cli"
	"github.com/grrtrr/clcv2/utils"
	"github.com/grrtrr/exit"
	"github.com/olekukonko/tablewriter"
)

func main() {
	var policies []clcv2.SBSServerPolicy
	var day = time.Now()
	var date = flag.String("start", day.Format("2006-01-02"), "Day to query storage usage for")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <Server-Policy-ID | Server-Name>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(0)
	}

	day, err := time.Parse("2006-01-02", *date)
	if err != nil {
		exit.Error("invalid backup query date %s (expected format: YYYY-MM-DD)", *date)
	}

	client, err := clcv2cli.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	// First look up the corresponding Server Policy or Policies, since the query needs the Account Policy ID.
	if utils.LooksLikeServerName(flag.Arg(0)) {
		policies, err = client.SBSgetServerPolicyDetails(flag.Arg(0))
		if err != nil {
			exit.Fatalf("failed to list SBS policies for server %s: %s", flag.Arg(0), err)
		}
	} else {
		p, err := client.SBSgetServerPolicy(flag.Arg(0))
		if err != nil {
			exit.Fatalf("failed to list SBS Server Policy %s: %s", flag.Arg(0), err)
		}
		policies = []clcv2.SBSServerPolicy{*p}
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetAlignment(tablewriter.ALIGN_CENTER)
	table.SetAutoWrapText(false)
	table.SetHeader([]string{"Server",
		fmt.Sprintf("Usage on %s", day.Format("Mon, Jan 2 2006")),
		"Server Policy ID", "Account Policy ID",
	})

	for _, p := range policies {
		usage, err := client.SBSgetServerStorageUsage(p.AccountPolicyID, p.ID, day)
		if err != nil {
			exit.Fatalf("failed to obtain server %s storage use on %s: %s", p.ServerID, day.Format("2006-01-02"), err)
		}
		table.Append([]string{p.ServerID, humanize.Bytes(usage), p.ID, p.AccountPolicyID})
	}
	table.Render()
}
