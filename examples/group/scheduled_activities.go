/*
 * Get the scheduled activities associated with a group.
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

	"github.com/dustin/go-humanize"
	"github.com/grrtrr/clcv2/clcv2cli"
	"github.com/grrtrr/exit"
	"github.com/kr/pretty"
	"github.com/olekukonko/tablewriter"
)

func main() {
	var uuid string
	var simple = flag.Bool("simple", false, "Use simple (debugging) output format")
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

	client, err := clcv2cli.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	if _, err := hex.DecodeString(flag.Arg(0)); err == nil {
		uuid = flag.Arg(0)
	} else if *location == "" {
		exit.Errorf("Need a location argument (-l) if not using Group UUID (%s)", flag.Arg(0))
	} else {
		if grp, err := client.GetGroupByName(flag.Arg(0), *location); err != nil {
			exit.Errorf("failed to resolve group name %q: %s", flag.Arg(0), err)
		} else if grp == nil {
			exit.Errorf("No group named %q was found in %s", flag.Arg(0), *location)
		} else {
			uuid = grp.Id
		}
	}

	sa, err := client.GetGroupScheduledActivities(uuid)
	if err != nil {
		exit.Fatalf("failed to query billing details of group %s: %s", flag.Arg(0), err)
	}

	if *simple {
		pretty.Println(sa)
	} else {

		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoFormatHeaders(false)
		table.SetAlignment(tablewriter.ALIGN_RIGHT)
		table.SetAutoWrapText(false)

		table.SetHeader([]string{"Type", "On?", "Exp?",
			"Repeat", "Expire", "Begin", "Next", "#Count", "Days", "Modified",
		})

		for _, s := range sa {
			var beginStr, nextStr string

			modifiedStr := humanize.Time(s.ChangeInfo.ModifiedDate)
			/* The ModifiedBy field can be an email address, or an API Key (hex string) */
			if _, err := hex.DecodeString(s.ChangeInfo.ModifiedBy); err == nil {
				modifiedStr += " via API Key"
			} else {
				modifiedStr += " by " + s.ChangeInfo.ModifiedBy
			}

			if s.BeginDateUtc.IsZero() {
				beginStr = "never"
			} else {
				beginStr = s.BeginDateUtc.In(time.Local).Format("Mon 2 Jan 15:04 MST")
			}
			if s.NextOccurrenceDateUtc.IsZero() {
				nextStr = "never"
			} else {
				nextStr = s.NextOccurrenceDateUtc.In(time.Local).Format("Mon 2 Jan 15:04 MST")
			}
			table.Append([]string{
				s.Type, s.Status, fmt.Sprint(s.IsExpired),
				s.Repeat, s.Expire, beginStr, nextStr,
				fmt.Sprint(s.OccurrenceCount),
				strings.Join(s.CustomWeeklyDays, "/"), modifiedStr,
			})

		}
		table.Render()
	}
}
