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

	humanize "github.com/dustin/go-humanize"
	"github.com/grrtrr/clcv2/clcv2cli"
	"github.com/grrtrr/exit"
	"github.com/kr/pretty"
	"github.com/olekukonko/tablewriter"
)

func main() {
	var simple = flag.Bool("simple", false, "Use simple (debugging) output format")
	flag.Parse()

	client, err := clcv2cli.NewCLIClient()
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

		table.SetHeader([]string{"CLC => Remote", "CLC Nets",
			"Remote", "Remote Nets", "ID", "Last Change",
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
				fmt.Sprintf("%s => %-15s", v.Local.LocationAlias, v.Remote.Address),
				strings.Join(v.Local.Subnets, ", "),
				v.Remote.SiteName,
				strings.Join(v.Remote.Subnets, ", "),
				v.ID,
				modifiedStr,
			})
		}
		table.Render()
	}
}
