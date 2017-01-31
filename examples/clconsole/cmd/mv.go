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
		Use:     "mv  [server|group [server|group...]]  <dest-group>",
		Aliases: []string{"move"},
		Short:   "Move server(s)/group(s) into different folder",
		Long:    "Move one or more groups/servers to a different folder",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return errors.Errorf("Need one or more sources (group/server) and a destination folder")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			var l = len(args)
			var servers []string
			var groups = make(map[string]string) // map { name -> id }
			var wg sync.WaitGroup

			for _, name := range args[:l-1] {
				if isServer, where, err := groupOrServer(name); err != nil {
					fmt.Fprintf(os.Stderr, "Error looking up %s: %s\n", name, err)
				} else if isServer { // moving a server
					servers = append(servers, where)
				} else {
					groups[name] = where
				}
			}

			// Destination folder (aka Hardware Group) is the last argument.
			isServer, newParent, err := groupOrServer(args[l-1])
			if err != nil {
				exit.Errorf("failed to look up destination group %s: %s", args[l-1], err)
			} else if isServer {
				exit.Errorf("destination %q does not look like a folder name", args[l-1])
			}

			for _, server := range servers {
				where := server
				wg.Add(1)
				go func() {
					// Note: this does not detect whether the server is already located under @newParent
					if err = client.ServerSetGroup(where, newParent); err != nil {
						fmt.Fprintf(os.Stderr, "Failed to change the parent group of server %s: %s\n", server, err)
					} else {
						fmt.Printf("Successfully moved %s to %s\n", server, args[l-1])
					}
					wg.Done()
				}()
			}
			for group, where := range groups {
				group, where := group, where
				wg.Add(1)
				go func() {
					if where == newParent {
						fmt.Printf("Nothing to do - source group %s is the same as destination.\n", group)
					} else if err = client.GroupSetParent(where, newParent); err != nil {
						fmt.Fprintf(os.Stderr, "Failed to change the parent of group %s: %s\n", group, err)
					} else {
						fmt.Printf("Successfully moved %s to %s\n", group, args[l-1])
					}
					wg.Done()
				}()
			}
			wg.Wait()
		},
	})
}
