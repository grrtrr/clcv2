package main

/*
 * Creates a Server Policy for a given Account Policy and Server.
 */
import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
	"github.com/olekukonko/tablewriter"
)

const (
	// The API allows a maximum date range of 100 days
	maxDateRange = 100
)

func main() {
	var region = flag.String("region", "", "Region where backups are stored, one of US EAST, US WEST, CANADA, GREAT BRITAIN, GERMANY, APAC")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <Account-Policy-ID> <Server-Name>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
		os.Exit(0)
	}
	flag.Parse()

	if flag.NArg() != 2 {
		flag.Usage()
	} else if *region == "" {
		fmt.Fprintf(os.Stderr, "Need a -region argument.\n")
		flag.Usage()
	}

	client, err := clcv2.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	p, err := client.SBScreateServerPolicy(flag.Arg(0), flag.Arg(1), *region)
	if err != nil {
		exit.Fatalf("failed to create policy for server %s: %s", flag.Arg(0), err)

	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoWrapText(false)
	table.SetHeader([]string{"Server", "New Server Policy ID", "Status", "Region", "Account",
		"Unsubscribe Date", "Expiration Date"})

	table.Append([]string{p.ServerID, p.ID, p.Status, p.StorageRegion, p.ClcAccountAlias,
		fmt.Sprint(p.UnsubscribedDate), fmt.Sprint(p.ExpirationDate)})
	table.Render()
}
