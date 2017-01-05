package cmd

import (
	"fmt"
	"os"

	"golang.org/x/sync/errgroup"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func init() {
	Root.AddCommand(&cobra.Command{
		Use:     "creds [group|server [group|server]...]",
		Aliases: []string{"credentials"},
		Short:   "Print login credentials of server(s)",
		Run: func(cmd *cobra.Command, args []string) {
			if servers, err := extractServerNames(args); err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
			} else if len(servers) > 0 {
				var eg errgroup.Group
				var table = tablewriter.NewWriter(os.Stdout)

				table.SetAutoFormatHeaders(false)
				table.SetAlignment(tablewriter.ALIGN_LEFT)
				table.SetAutoWrapText(true)
				table.SetHeader([]string{"VM", "Username", "Password"})

				for _, name := range servers {
					eg.Go(func() error {
						if creds, err := client.GetServerCredentials(name); err != nil {
							fmt.Fprintf(os.Stderr, "ERROR (%s): %s\n", name, err)
						} else {
							table.Append([]string{name, creds.Username, creds.Password})
						}
						return err
					})
				}
				_ = eg.Wait()
				table.Render()
			}
		},
	})
}
