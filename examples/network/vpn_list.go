/*
 * List all the Site-to-Site VPNs of the given account.
 */
package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"os"
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
	flag.Parse()

	client, err := clcv2.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	vpns, err := client.GetVPNs()
	if err != nil {
		exit.Fatalf("failed to list VPNs: %s", err)
	}

	if len(vpns) == 0 {
		fmt.Printf("No VPNs registered for account %s.\n", client.AccountAlias)
	} else if *simple {
		pretty.Println(vpns)
	} else {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoFormatHeaders(false)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetAutoWrapText(true)

		table.SetHeader([]string{
			"Local", "Local Nets", "Remote", "Remote Nets",
			"IKE", "IP Sec",
			"ID", "Last Change",
		})

		for _, v := range vpns {
			modifiedStr := humanize.Time(v.ChangeInfo.ModifiedDate)
			// The ModifiedBy field can be an email address, or an API Key (hex string)
			if _, err := hex.DecodeString(v.ChangeInfo.ModifiedBy); err == nil {
				modifiedStr += " via API Key"
			} else if len(v.ChangeInfo.ModifiedBy) > 6 {
				modifiedStr += " by " + v.ChangeInfo.ModifiedBy[:6]
			} else {
				modifiedStr += " by " + v.ChangeInfo.ModifiedBy
			}

			table.Append([]string{
				fmt.Sprintf("%s: %s", v.Local.LocationAlias, v.Local.Address),
				strings.Join(v.Local.Subnets, ", "),
				fmt.Sprintf("%s: %s", v.Remote.SiteName, v.Remote.Address),
				strings.Join(v.Remote.Subnets, ", "),
				fmt.Sprintf("%s %s/%s, %s, NAT: %t", v.IKE.DiffieHellmanGroup, v.IKE.Encryption, v.IKE.Hashing,
					time.Duration(v.IKE.Lifetime)*time.Second, v.IKE.NatTraversal),
				fmt.Sprintf("%s %s/%s, %s", v.IPsec.Protocol, v.IPsec.Encryption, v.IPsec.Hashing,
					time.Duration(v.IPsec.Lifetime)*time.Second),
				v.ID,
				modifiedStr,
			})
		}
		table.Render()
	}

}
