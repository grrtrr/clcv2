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
	var start = flag.String("start", startTime.Format("2006-01-02"), "Start date of the backup")
	var end = flag.String("end", "", "End date of the backup")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <account-policy-ID> <server-policy-ID\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(0)
	}

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

	restorePoints, err := client.SBSgetRestorePointDetails(flag.Arg(0), flag.Arg(1), startTime, endTime)
	if err != nil {
		exit.Fatalf("failed to list SBS restore point details found for server policy %s: %s", flag.Arg(1), err)

	}

	if len(restorePoints) == 0 {
		fmt.Printf("No restore point details found found for Server Policy %s between %s ... %s.\n",
			flag.Arg(1), startTime.Format("Mon, 2 Jan 2006"), endTime.Format("Mon, 2 Jan 2006"))
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