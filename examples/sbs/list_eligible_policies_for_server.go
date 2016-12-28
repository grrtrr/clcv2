package main

/*
 * Print the list of Account Policies eligible for the specified server.
 */
import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/grrtrr/clcv2/clcv2cli"
	"github.com/grrtrr/exit"
	"github.com/olekukonko/tablewriter"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <server-name>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(0)
	}

	client, err := clcv2cli.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	policies, err := client.SBSgetEligiblePolicies(flag.Arg(0))
	if err != nil {
		exit.Fatalf("failed to list eligible policies for %s: %s", flag.Arg(0), err)
	}

	if len(policies) == 0 {
		fmt.Printf("No eligible policies found for %s.\n", flag.Arg(0))
	} else {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoFormatHeaders(false)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetAutoWrapText(false)

		table.SetHeader([]string{"Name", "Policy ID", "OS", "Status", "Freq/h", "Ret/d", "Paths"})

		for _, p := range policies {
			table.Append([]string{p.Name, p.PolicyID, p.OsType, p.Status, fmt.Sprint(p.BackupIntervalHours),
				fmt.Sprint(p.RetentionDays), strings.Join(p.Paths, ", ")})
		}
		table.Render()
	}
}
