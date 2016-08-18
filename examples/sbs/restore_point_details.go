package main

/*
 * List the restore point details for a given Account and Server Policy.
 */
import (
	"flag"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/clcv2/utils"
	"github.com/grrtrr/exit"
	"github.com/kr/pretty"
)

const (
	// The API allows a maximum date range of 100 days
	maxDateRange = 100
)

func main() {
	var startTime = time.Now().AddDate(0, 0, -maxDateRange)
	var endTime time.Time
	var start = flag.String("start", startTime.Format("2006-01-02"), "Start date of the backup (format YYYY-MM-DD)")
	var end = flag.String("end", "", "End date of the backup (format YYYY-MM-DD)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <server-policy-ID\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(0)
	}
	fmt.Println(flag.Arg(0), utils.LooksLikeServerName(flag.Arg(0)))
	return
	// Date range sanity checks
	startTime, err := time.Parse("2006-01-02", *start)
	if err != nil {
		exit.Error("invalid backup start date %s (expected format: YYYY-MM-DD)", *start)
	}

	if *end == "" {
		endTime = startTime.AddDate(0, 0, maxDateRange)
		if endTime.After(time.Now()) {
			fmt.Println(endTime, "later than NOW")
			endTime = time.Now()
		}
	} else if endTime, err = time.Parse("2006-01-02", *end); err != nil {
		exit.Error("invalid backup end date %s (expected format: YYYY-MM-DD)", *end)
	}

	client, err := clcv2.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	// Look up the Account Policy ID associated with the Server Policy ID
	p, err := client.SBSgetServerPolicy(flag.Arg(0))
	if err != nil {
		exit.Fatalf("failed to look up Account Policy of %s: %s.", flag.Arg(0), err)
	}

	restorePoints, err := client.SBSgetRestorePointDetails(flag.Arg(0), p.AccountPolicyID, startTime, endTime)
	if err != nil {
		exit.Fatalf("failed to list SBS restore point details found for server policy %s: %s", flag.Arg(0), err)

	}

	if len(restorePoints) == 0 {
		fmt.Printf("No restore point details found found for Server Policy %s between %s ... %s.\n",
			flag.Arg(0), startTime.Format("Mon, 2 Jan 2006"), endTime.Format("Mon, 2 Jan 2006"))
	} else {
		pretty.Println(restorePoints)
		/*
			table := tablewriter.NewWriter(os.Stdout)
			table.SetAutoFormatHeaders(false)
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.SetAutoWrapText(false)
			table.SetHeader([]string{"Server", "Server Policy ID", "Status", "Region", "Account",
				"Unsubscribe Date", "Expiration Date"})

			for _, p := range policies {
				table.Append([]string{p.ServerID, p.ID, p.Status, p.StorageRegion, p.ClcAccountAlias,
					fmt.Sprint(p.UnsubscribedDate), fmt.Sprint(p.ExpirationDate)})
			}
			table.Render()
		*/
	}
}
