/*
 * Get the current and estimated charges for each server in a designated group hierarchy.
 */
package main

import (
	"github.com/olekukonko/tablewriter"
	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
	"github.com/kr/pretty"
	"encoding/hex"
	"strings"
	"path"
	"flag"
	"fmt"
	"os"
)

func main() {
	var uuid string
	var simple   = flag.Bool("simple", false, "Use simple (debugging) output format")
	var location = flag.String("l", "", "Location to use if using a Group-Name instead of a UUID")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <Group Name or UUID>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	client, err := clcv2.NewClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	if _, err := hex.DecodeString(flag.Arg(0)); err == nil {
		uuid = flag.Arg(0)
	} else if  *location == "" {
		exit.Errorf("Need a location argument (-l) if not using Group UUID (%s)", flag.Arg(0))
	} else {
		if grp, err := client.GetGroupByName(flag.Arg(0), *location); err != nil {
			exit.Errorf("Failed to resolve group name %q: %s", flag.Arg(0), err)
		} else if grp == nil {
			exit.Errorf("No group named %q was found on %s", flag.Arg(0), *location)
		} else {
			uuid = grp.Id
		}
	}

	bd, err := client.GetGroupBillingDetails(uuid)
	if err != nil {
		exit.Fatalf("Failed to query billing details of group %s: %s", flag.Arg(0), err)
	}

	if *simple {
		pretty.Println(bd)
	} else {
		var totalMonthTod, totalMonthEst float64

		if rootGroup, ok := bd.Groups[uuid]; !ok {
			exit.Fatalf("Query result does not contain queried group %s", uuid)
		} else {
			fmt.Printf("Billing details of %s as of %s:\n", rootGroup.Name,
				   bd.Date.Format("Monday, 2 Jan 2006, 15:04 MST"))
		}

		for k, v := range bd.Groups {
			if k == uuid || len(v.Servers) == 0 {
				continue
			}
			fmt.Printf("\nServer details of %q (%s):\n", v.Name, k)

			table := tablewriter.NewWriter(os.Stdout)
			table.SetAutoFormatHeaders(false)
			table.SetAlignment(tablewriter.ALIGN_RIGHT)
			table.SetAutoWrapText(true)

			table.SetHeader([]string{ "Server", "Template Cost", "Archive Cost", "Current Hour",
						  "Month to Date", "Monthly Estimate" })

			for s, sbd := range v.Servers {
				table.Append([]string{ strings.ToUpper(s),
					fmt.Sprintf("$%.2f", sbd.TemplateCost),
					fmt.Sprintf("$%.2f", sbd.ArchiveCost),
					fmt.Sprintf("$%.2f", sbd.CurrentHour),
					fmt.Sprintf("$%.2f", sbd.MonthToDate),
					fmt.Sprintf("$%.2f", sbd.MonthlyEstimate),
				})
				totalMonthTod += sbd.MonthToDate
				totalMonthEst += sbd.MonthlyEstimate
			}
			table.Render()
		}

		fmt.Printf("\nTotal month to date:    $%.2f\n", totalMonthTod)
		fmt.Printf("Total monthly estimate: $%.2f\n", totalMonthEst)
	}
}
