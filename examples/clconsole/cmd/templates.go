package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/grrtrr/exit"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func init() {
	Root.AddCommand(&cobra.Command{
		Use:     "templates  [location]",
		Aliases: []string{"net"},
		Short:   "List available templates",
		Long:    "List templates available in a given region. If @location argument is present, it overrides the default region.",
		Run: func(cmd *cobra.Command, args []string) {
			var region = location

			if len(args) > 0 {
				region = args[0]
			}

			capa, err := client.GetDeploymentCapabilities(region)
			if err != nil {
				exit.Fatalf("failed to query deployment capabilities of %s: %s", region, err)
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetAutoFormatHeaders(false)
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.SetAutoWrapText(false)

			/* Note: not displaying ReservedDrivePaths and DrivePathLength here, I don't understand their use. */
			/* Note: not listing Capabilities here, since the table gets too large for a single screen */
			table.SetHeader([]string{
				fmt.Sprintf("%s Template Name", strings.ToUpper(region)), "Description", "OS", "Storage"})

			for _, tpl := range capa.Templates {
				table.Append([]string{tpl.Name, tpl.Description, tpl.OsType, fmt.Sprintf("%d GB", tpl.StorageSizeGB)})
			}
			table.Render()
		},
	})
}
