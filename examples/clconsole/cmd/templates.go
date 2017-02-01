package cmd

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var templateFlags struct {
	csv bool // Whether to output in CSV format
	sep Rune // Separator for @csv output
}

func init() {
	templateFlags.sep.Set(",")
	ShowTemplates.Flags().BoolVar(&templateFlags.csv, "csv", false, "Whether to output in CSV format, or a pretty-printed table")
	ShowTemplates.Flags().VarP(&templateFlags.sep, "sep", "s", "Field separator for use with CSV option")

	Root.AddCommand(ShowTemplates)
}

var ShowTemplates = &cobra.Command{
	Use:     "templates  [location]",
	Aliases: []string{"templ", "tp"},
	Short:   "List available templates",
	Long:    "List templates available in a given region. If @location argument is present, it overrides the default region.",
	Run: func(cmd *cobra.Command, args []string) {
		var (
			region       = location                            // global flag
			templateData = func(tpl clcv2.Template) []string { // print a single table/CSV row
				return []string{tpl.Name, tpl.Description, tpl.OsType, fmt.Sprintf("%d GB", tpl.StorageSizeGB)}
			}
		)

		if len(args) > 0 {
			region = args[0]
		} else if region == "" {
			region = client.LocationAlias
		}

		capa, err := client.GetDeploymentCapabilities(region)
		if err != nil {
			exit.Fatalf("failed to query deployment capabilities of %s: %s", region, err)
		}

		/* Note: not displaying ReservedDrivePaths and DrivePathLength here, I don't understand their use. */
		/* Note: not listing Capabilities here, since the table gets too large for a single screen */
		header := []string{fmt.Sprintf("%s Template Name", strings.ToUpper(region)), "Description", "OS", "Storage"}
		if templateFlags.csv {
			var w = csv.NewWriter(os.Stdout)

			w.Comma = rune(templateFlags.sep)
			if err := w.Write(header); err != nil {
				exit.Fatalf("failed to write CSV header: %s", err)
			}

			for _, tpl := range capa.Templates {
				if err := w.Write(templateData(tpl)); err != nil {
					exit.Fatalf("failed to write CSV row data: %s", err)
				}
			}
			w.Flush()
			if err := w.Error(); err != nil {
				exit.Fatalf("failed to write CSV template data: %s", err)
			}
		} else { // pretty table
			var table = tablewriter.NewWriter(os.Stdout)

			table.SetAutoFormatHeaders(false)
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.SetAutoWrapText(false)

			table.SetHeader(header)
			for _, tpl := range capa.Templates {
				table.Append(templateData(tpl))
			}
			table.Render()
		}
	},
}
