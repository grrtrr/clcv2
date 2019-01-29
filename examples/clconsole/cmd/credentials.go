package cmd

import (
	"fmt"
	"os"

	"github.com/grrtrr/clcv2"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

func init() {
	Root.AddCommand(&cobra.Command{
		Use:     "creds [group|server [group|server]...]",
		Aliases: []string{"credentials"},
		Short:   "Print login credentials of server(s)",
		PreRunE: checkAtLeastArgs(1, "Need at least 1 group or server name"),
		Run: func(cmd *cobra.Command, args []string) {
			if servers, err := extractServerNames(args); err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
			} else if len(servers) > 0 {
				showServerCredentials(client, servers...)
			}
		},
	})
}

// showServerCredentials displays the login credentials of one or more @servers.
func showServerCredentials(client *clcv2.CLIClient, servers ...string) {
	if len(servers) > 0 {
		var eg errgroup.Group
		var numEntries int

		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoFormatHeaders(false)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetAutoWrapText(true)
		table.SetHeader([]string{"VM", "Username", "Password"})

		for _, name := range servers {
			name := name
			eg.Go(func() error {
				creds, err := client.GetServerCredentials(name)
				if err != nil {
					fmt.Fprintf(os.Stderr, "ERROR (%s): %s\n", name, err)
					return err
				}
				table.Append([]string{name, creds.Username, creds.Password})
				numEntries++
				return nil
			})
		}

		if err := eg.Wait(); err != nil {
			die("failed to get credentials: %s", err)
		} else if numEntries > 0 {
			table.Render()
		} else {
			fmt.Println("No credentials available.")
		}
	}
}
