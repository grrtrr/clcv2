package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/grrtrr/clcv2"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

var Delete = &cobra.Command{
	Use:     "delete  [group|server [group|server]...]",
	Aliases: []string{"remove", "rm", "del"},
	Short:   "Delete server(s)/group(s) (CAUTION)",
	Long:    "Completely and irreversibly removes servers/groups - USE WITH CAUTION",
	RunE: func(cmd *cobra.Command, args []string) error {
		var eg errgroup.Group

		for _, name := range args {
			name := name
			eg.Go(func() error {
				var reqID string

				isServer, where, err := groupOrServer(name)
				if err != nil {
					fmt.Fprintf(os.Stderr, "ERROR (%s): %s\n", name, err)
				} else if isServer {
					if reqID, err = client.DeleteServer(where); err != nil {
						fmt.Fprintf(os.Stderr, "ERROR deleting server %s: %s\n", name, err)
					} else {
						log.Printf("Deleting server %s: %s", name, reqID)
					}
				} else {
					if reqID, err = client.DeleteGroup(where); err != nil {
						fmt.Fprintf(os.Stderr, "ERROR deleting group %s: %s\n", name, err)
					} else {
						log.Printf("Deleting group %s: %s\n", name, reqID)
					}
				}

				if reqID != "" {
					client.PollStatusFn(reqID, intvl, func(s clcv2.QueueStatus) {
						log.Printf("Deleting %s: %s", name, s)
					})
				}
				return err
			})
		}
		_ = eg.Wait()
		return nil
	},
}

func init() {
	Root.AddCommand(Delete)
}
