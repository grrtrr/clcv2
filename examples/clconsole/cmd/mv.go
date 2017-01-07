package cmd

import (
	"fmt"
	"os"
	"sync"

	"github.com/grrtrr/exit"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	Root.AddCommand(&cobra.Command{
		Use:   "mv  [server|group [server|group...]]  <dest-group>",
		Short: "Move server(s)/group(s) into different folder",
		Long:  "Move one or more groups/servers to a different folder",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return errors.Errorf("Need one or more sources (group/server) and a destination folder")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			var l = len(args)
			var wg sync.WaitGroup

			isServer, newParent, err := groupOrServer(args[l-1])
			if err != nil {
				exit.Errorf("failed to look up destination group %s: %s", args[l-1], err)
			} else if isServer {
				exit.Errorf("destination %q does not look like a folder name", args[l-1])
			}

			for _, name := range args[:l-1] {
				name := name
				wg.Add(1)
				go func() {
					if isServer, where, err := groupOrServer(name); err != nil {
						fmt.Fprintf(os.Stderr, "Error looking up %s: %s\n", name, err)
					} else if isServer { // moving a server
						if err = client.ServerSetGroup(where, newParent); err != nil {
							fmt.Fprintf(os.Stderr, "Failed to change the parent group of server %s: %s\n", name, err)
						} else {
							fmt.Printf("Successfully moved %s to %s\n", name, args[l-1])
						}
					} else if where == newParent {
						fmt.Printf("Nothing to do - source group %s is the same as destination.\n", name)
					} else if err = client.GroupSetParent(where, newParent); err != nil {
						fmt.Fprintf(os.Stderr, "Failed to change the parent of group %s: %s\n", name, err)
					} else {
						fmt.Printf("Successfully moved %s to %s\n", name, args[l-1])
					}
					wg.Done()
				}()
			}
			wg.Wait()
		},
	})
}
