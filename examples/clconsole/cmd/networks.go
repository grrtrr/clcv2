package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func init() {
	Root.AddCommand(&cobra.Command{
		Use:     "networks  [location]",
		Aliases: []string{"nets", "net"},
		Short:   "Show available networks",
		Long:    "Show networks available to current account. If @location argument is present, it overrides the default data centre (region)k.",
		Run: func(cmd *cobra.Command, args []string) {
			var region = location

			if len(args) > 0 {
				region = args[0]
			}

			fmt.Printf("Networks visible to %s account in %s:\n", client.AccountAlias, strings.ToUpper(region))
			showNetworks(client, region, client.AccountAlias)
			if client.AccountAlias != client.RegisteredAccountAlias() {
				fmt.Printf("Networks visible to parent %s account:\n", client.RegisteredAccountAlias())
				showNetworks(client, region, client.RegisteredAccountAlias())
			}
		},
	})
}

// showNetworks shows networks visible to @account in data centre location @location.
func showNetworks(client *clcv2.CLIClient, location, account string) {
	networks, err := client.GetNetworks(location, account)
	if err != nil {
		exit.Fatalf("failed to list networks in %s: %s", location, err)
	}

	if len(networks) == 0 {
		println("Empty result.")
	} else {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoFormatHeaders(false)
		table.SetAlignment(tablewriter.ALIGN_RIGHT)
		table.SetAutoWrapText(false)

		table.SetHeader([]string{"CIDR", "Gateway", "VLAN", "Name", "Description", "Type", "ID"})
		for _, l := range networks {
			table.Append([]string{l.Cidr, l.Gateway, fmt.Sprint(l.Vlan), l.Name, l.Description, l.Type, l.Id})
		}
		table.Render()
	}
}
