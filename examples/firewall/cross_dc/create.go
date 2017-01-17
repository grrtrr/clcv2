/*
 * Create a new cross-datacenter firewall policy.
 */
package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/clcv2/clcv2cli"
	"github.com/grrtrr/exit"
	"github.com/olekukonko/tablewriter"
)

func main() {
	var (
		srcNet = flag.String("src", "", "Source Network in CIDR notation")
		dstNet = flag.String("dst", "", "Destination Network in CIDR notation")
		acct   = flag.String("da", "", "Destination account (defaults to source account)")
		dstLoc = flag.String("dc", "", "Destination data centre alias")
		enable = flag.Bool("enable", true, "Whether the firewall policy is initially enabled")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options] <Location>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(0)
	}

	client, err := clcv2cli.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	if *acct == "" {
		*acct = client.AccountAlias
	}

	req := clcv2.CrossDataCenterFirewallPolicyReq{
		SourceCIDR:   *srcNet,
		DestCIDR:     *dstNet,
		DestAccount:  *acct,
		DestLocation: *dstLoc,
		Enabled:      *enable,
	}

	p, err := client.CreateCrossDataCenterFirewallPolicy(flag.Arg(0), &req)
	if err != nil {
		exit.Fatalf("failed to create cross-datacenter firewall policy in %s: %s", flag.Arg(0), err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetAlignment(tablewriter.ALIGN_CENTER)
	table.SetAutoWrapText(false)

	table.SetHeader([]string{"Data Center", "Network", "Account", "Status", "ID"})

	enabledStr := "*"
	if !p.Enabled {
		enabledStr = "-"
	}
	table.Append([]string{
		strings.ToUpper(fmt.Sprintf("%s => %s", p.SourceLocation, p.DestLocation)),
		fmt.Sprintf("%18s => %-18s", p.SourceCIDR, p.DestCIDR),
		strings.ToUpper(fmt.Sprintf("%-4s => %4s", p.SourceAccount, p.DestAccount)),
		fmt.Sprintf("%s%s", enabledStr, p.Status),
		p.ID,
	})

	table.Render()
}
