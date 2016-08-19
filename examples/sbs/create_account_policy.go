/*
 * Create a new account policy.
 */
package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
	"github.com/olekukonko/tablewriter"
)

func main() {
	var osType = flag.String("os", "Linux", "The OS type (only supported values are 'Linux' and 'Windows')")
	var paths = flag.String("paths", "/home,/etc,/opt", "Comma-separated list of paths to include")
	var exclude = flag.String("exclude", "/opt/simplebackupservice", "Comma-separated list of paths to exclude (e.g. /tmp,/var/run)")
	var intvl = flag.Duration("freq", 6*time.Hour, "Backup interval (time duration between backups")
	var keep = flag.Int("keep", 10, "The number of days backup data will be retained")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options] <Account-Policy-Name>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(0)
	}

	client, err := clcv2.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	req := clcv2.SBSAccountPolicy{
		// The backup frequency of the Policy (= duration between backups)
		BackupIntervalHours: int(intvl.Hours()),

		// The account alias that the Policy belongs to
		ClcAccountAlias: client.AccountAlias,

		// A list of the directories that the Policy excludes from backup
		ExcludedDirectoryPaths: parseCSV(*exclude),

		// The name of the Policy
		Name: flag.Arg(0),

		// The OS Type - 'Linux' or 'Windows'
		OsType: *osType,

		// A list of the directories that the Policy includes in each backup
		Paths: parseCSV(*paths),

		// The number of days backup data will be retained
		RetentionDays: *keep,
	}

	p, err := client.SBScreatePolicy(&req)
	if err != nil {
		exit.Fatalf("failed to create new SBS account policy %s: %s", flag.Arg(0), err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoWrapText(false)
	table.SetHeader([]string{"Name", "Policy ID", "OS", "Status", "Freq/h", "Ret/d", "Paths"})

	table.Append([]string{p.Name, p.PolicyID, p.OsType, p.Status, fmt.Sprint(p.BackupIntervalHours),
		fmt.Sprint(p.RetentionDays), strings.Join(p.Paths, ", ")})
	table.Render()
}

// parseCSV splits @s as comma-separated list of values, removing leading/trailing whitespace from elements
func parseCSV(s string) []string {
	var res []string

	for _, val := range strings.Split(s, ",") {
		res = append(res, strings.TrimSpace(val))
	}
	return res
}
