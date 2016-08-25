/*
 * Print details of a single VPN.
 */
package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
	"github.com/kr/pretty"
	"github.com/olekukonko/tablewriter"
)

func main() {
	var simple = flag.Bool("simple", false, "Use simple (debugging) output format")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <VPN ID>\n", path.Base(os.Args[0]))
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

	vpn, err := client.GetVPN(flag.Arg(0))
	if err != nil {
		exit.Fatalf("failed to query VPN %s: %s", flag.Arg(0), err)
	}

	if *simple {
		pretty.Println(vpn)
	} else {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoFormatHeaders(false)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetAutoWrapText(false)

		//		table.SetHeader([]string{"CLC => Remote", "CLC Nets",			"Remote", "Remote Nets", "ID", "Last Change",		})

		modifiedStr := humanize.Time(vpn.ChangeInfo.ModifiedDate)
		// The ModifiedBy field can be an email address, or an API Key (hex string)
		if _, err := hex.DecodeString(vpn.ChangeInfo.ModifiedBy); err == nil {
			modifiedStr += " via API Key"
		} else {
			modifiedStr += ", by " + vpn.ChangeInfo.ModifiedBy
		}

		table.Append([]string{fmt.Sprintf("%s subnets:", vpn.Local.LocationAlias), strings.Join(vpn.Local.Subnets, ", ")})
		table.Append([]string{fmt.Sprintf("%s endpoint:", vpn.Local.LocationAlias), vpn.Local.Address})
		table.Append([]string{fmt.Sprintf("%s endpoint:", vpn.Remote.SiteName), vpn.Remote.Address})
		table.Append([]string{fmt.Sprintf("%s subnets:", vpn.Remote.SiteName), strings.Join(vpn.Remote.Subnets, ", ")})
		table.Append([]string{"IKE configuration:", fmt.Sprintf("%s, %s/%s, Key: %q, NAT traversal: %t, Lifetime: %s",
			vpn.IKE.DiffieHellmanGroup,
			vpn.IKE.Encryption, vpn.IKE.Hashing, vpn.IKE.PreSharedKey,
			vpn.IKE.NatTraversal,
			time.Duration(vpn.IKE.Lifetime)*time.Second)})
		table.Append([]string{"IPsec configuration:", fmt.Sprintf("%s, %s/%s, %s, Lifetime: %s",
			vpn.IPsec.Pfs, vpn.IPsec.Encryption, vpn.IPsec.Hashing,
			strings.ToUpper(vpn.IPsec.Protocol), time.Duration(vpn.IPsec.Lifetime)*time.Second)})
		table.Append([]string{"Last change:", modifiedStr})

		table.Render()
	}
}
