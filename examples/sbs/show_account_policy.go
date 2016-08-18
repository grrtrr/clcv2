/*
 * Print details of a single Account Policy given its account-policy-ID.
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

	p, err := client.SBSgetPolicy(flag.Arg(0))
	if err != nil {
		exit.Fatalf("failed to list SBS account policy %s: %s", flag.Arg(0), err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoWrapText(false)
	table.SetHeader([]string{"Name", "Policy ID", "OS", "Status", "Freq/h", "Ret/d", "Paths"})

	table.Append([]string{p.Name, p.PolicyID, p.OsType, p.Status, fmt.Sprint(p.BackupIntervalHours),
		fmt.Sprint(p.RetentionDays), strings.Join(p.Paths, ", ")})
	table.Render()
}
