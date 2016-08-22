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

	humanize "github.com/dustin/go-humanize"
	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
	"github.com/olekukonko/tablewriter"
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
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <server-policy-ID>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() != 1 {
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

	restorePoints, err := client.SBSgetRestorePointDetails(p.AccountPolicyID, flag.Arg(0), startTime, endTime)
	if err != nil {
		exit.Fatalf("failed to list SBS restore point details found for server policy %s: %s", flag.Arg(0), err)
	}

	if len(restorePoints) == 0 {
		fmt.Printf("No restore point details found found for server %s between %s and %s.\n", p.ServerID,
			startTime.Format("Mon, 2 Jan 2006"), endTime.Format("Mon, 2 Jan 2006"))
	} else {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoFormatHeaders(false)
		table.SetAlignment(tablewriter.ALIGN_RIGHT)
		table.SetAutoWrapText(false)
		table.SetHeader([]string{"Start", "Duration", "Status",
			"Transferred", "Failed", "Removed", "Unchanged",
			"Total #Files", "Expires",
		})

		for _, r := range restorePoints {
			var duration = r.BackupFinishedDate.Sub(r.BackupStartedDate)
			var runtime = duration.String()

			// If the backup did not finish (yet), the FinishedDate is set to a Unix Epoch date in the past.
			if duration < 0 {
				runtime = "unknown"
			}

			table.Append([]string{
				r.BackupStartedDate.Local().Format("Mon, 15:04 MST"), runtime,
				r.RestorePointCreationStatus,
				fmtTransfer(r.FilesTransferredToStorage, r.BytesTransferredToStorage),
				fmtTransfer(r.FilesFailedTransferToStorage, r.BytesFailedToTransfer),
				fmtTransfer(r.FilesRemovedFromDisk, r.BytesInStorageForItemsRemoved),
				fmtTransfer(r.UnchangedFilesNotTransferred, r.UnchangedBytesInStorage),
				fmt.Sprint(r.NumberOfProtectedFiles),
				r.RetentionExpiredDate.Local().Format("Jan 2/2006"),
			})
		}
		table.Render()
	}
}

// pretty-print #files/#bytes transfer statistics
func fmtTransfer(numFiles, numBytes uint64) string {
	var s string

	if numFiles == 0 && numBytes == 0 {
		return "none"
	}
	s = fmt.Sprintf("%d file", numFiles)
	if numFiles != 1 {
		s += "s"
	}
	return fmt.Sprintf("%s, %s", s, humanize.Bytes(numBytes))
}
