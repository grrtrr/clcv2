package cmd

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/grrtrr/clcv2"
	"github.com/spf13/cobra"
)

var Snapshot = &cobra.Command{
	Use:     "snapshot server(s)",
	Aliases: []string{"snap"},
	Short:   "Snapshots individual servers and/or servers contained in a group folder",
	Run: func(cmd *cobra.Command, args []string) {
		if servers, err := extractServerNames(args); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		} else {
			var wg sync.WaitGroup

			for _, name := range servers {
				name := name
				wg.Add(1)
				go func() {
					defer wg.Done()

					if reqID, err := client.SnapshotServer(name); err != nil {
						fmt.Fprintf(os.Stderr, "ERROR snapshotting server %s: %s\n", name, err)
					} else {
						log.Printf("%s snapshot: %s", name, reqID)

						client.PollStatusFn(reqID, intvl, func(s clcv2.QueueStatus) {
							log.Printf("%s snapshot: %s", name, s)
						})
					}
				}()
			}

			wg.Wait()
		}
	},
}

func init() {
	Root.AddCommand(Snapshot)
}
