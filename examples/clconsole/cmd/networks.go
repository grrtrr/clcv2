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
		Use:     "nets  [location]",
		Aliases: []string{"networks", "net"},
		Short:   "Show available networks",
		Long:    "Show networks available to current account. If @location argument is present, it overrides the default data centre (region).",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 {
				conf.Location = args[0]
			}

			fmt.Printf("Networks visible to %s account in %s:\n", client.AccountAlias, strings.ToUpper(conf.Location))
			showNetworks(client, conf.Location, client.AccountAlias)
			if client.AccountAlias != client.RegisteredAccountAlias() {
				fmt.Printf("Networks visible to parent account %s:\n", client.RegisteredAccountAlias())
				showNetworks(client, conf.Location, client.RegisteredAccountAlias())
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
		table.SetAutoWrapText(true)

		table.SetHeader([]string{"CIDR", "Gateway", "VLAN", "Name", "Description", "Type", "ID"})
		for _, l := range networks {
			table.Append([]string{l.Cidr, l.Gateway, fmt.Sprint(l.Vlan), l.Name, l.Description, l.Type, l.Id})
		}
		table.Render()
	}
}
