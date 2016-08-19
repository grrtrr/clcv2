/*
 * Update an existing Account Policy.
 */
package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
	"github.com/olekukonko/tablewriter"
)

func main() {
	var intvl = flag.Duration("freq", 0, "Backup interval (time duration between backups")
	var exclude = flag.String("exclude", "", "Comma-separated list of paths to exclude (e.g. /tmp,/var/run)")
	var name = flag.String("name", "", "New name of the Account Policy")
	var osType = flag.String("os", "", "The OS type (only supported values are 'Linux' and 'Windows')")
	var paths = flag.String("paths", "", "Comma-separated list of paths to include")
	var keep = flag.Int("keep", 0, "The number of days backup data will be retained")
	var status = flag.String("status", "", "Account Policy status ('ACITVE' or 'INACTIVE')")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options] <Account-Policy-ID>\n", path.Base(os.Args[0]))
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

	p, err := client.SBSgetPolicy(flag.Arg(0))
	if err != nil {
		exit.Fatalf("failed to retrieve Account Policy %s: %s", flag.Arg(0), err)
	}

	// Update policy values
	if *intvl != 0 {
		p.BackupIntervalHours = int(intvl.Hours())
	}
	if *exclude != "" {
		p.ExcludedDirectoryPaths = parseCSV(*exclude)
	}
	if *name != "" {
		p.Name = *name
	}
	if *osType != "" {
		p.OsType = *osType
	}
	if *paths != "" {
		p.Paths = parseCSV(*paths)
	}
	if *keep != 0 {
		p.RetentionDays = *keep
	}
	if *status != "" {
		p.Status = *status
	}

	p, err = client.SBSupdatePolicy(flag.Arg(0), &p)
	if err != nil {
		exit.Fatalf("failed to create update SBS account policy %s: %s", flag.Arg(0), err)
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
