/*
 * Print the list of SBS policies associated with the user's account.
 */
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
	"github.com/olekukonko/tablewriter"
)

func main() {
	flag.Parse()

	client, err := clcv2.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	res, err := client.SBSgetPolicies()
	if err != nil {
		exit.Fatalf("failed to list SBS policies: %s", err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoWrapText(false)
	table.SetHeader([]string{"Name", "Policy ID", "OS", "Status", "Freq/h", "Ret/d", "Paths"})

	for _, p := range res.Results {
		table.Append([]string{p.Name, p.PolicyID, p.OsType, p.Status, fmt.Sprint(p.BackupIntervalHours),
			fmt.Sprint(p.RetentionDays), strings.Join(p.Paths, ", ")})
	}
	table.Render()
}
