package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

func init() {
	Root.AddCommand(&cobra.Command{
		Use:   "archive  [group|server [group|server]...]",
		Short: "Archive server(s)",
		Long:  "Place server(s) into the special 'Archive' folder in the data centre",
		Run: func(cmd *cobra.Command, args []string) {
			var eg errgroup.Group

			for _, name := range args {
				name := name
				eg.Go(func() error {
					var reqID string

					isServer, where, err := groupOrServer(name)
					if err != nil {
						fmt.Fprintf(os.Stderr, "ERROR (%s): %s\n", name, err)
					} else if isServer {
						if reqID, err = client.ArchiveServer(where); err != nil {
							fmt.Fprintf(os.Stderr, "ERROR archiving server %s: %s\n", name, err)
						} else {
							log.Printf("Archiving server %s: %s", name, reqID)
						}
					} else {
						if reqID, err = client.ArchiveGroup(where); err != nil {
							fmt.Fprintf(os.Stderr, "ERROR archiving group %s: %s\n", name, err)
						} else {
							log.Printf("Archiving group %s: %s\n", name, reqID)
						}
					}

					if reqID != "" {
						client.PollStatusFn(reqID, intvl, func(s clcv2.QueueStatus) {
							log.Printf("Archiving %s: %s", name, s)
						})
					}
					return err
				})
			}
			_ = eg.Wait()
		},
	})

	Root.AddCommand(&cobra.Command{
		Use:     "restore  <group|server>  <destination folder>",
		Short:   "Restore server/group into destination folder",
		Long:    "Restore server/group from the special CLC 'Archive' folder into designated destination folder",
		PreRunE: checkArgs(2, "Need server/group name and a destination folder for the restoration"),
		Run: func(cmd *cobra.Command, args []string) {
			var reqID, dest string

			isServer, dest, err := groupOrServer(args[1])
			if err != nil {
				exit.Errorf("failed to resolve destination folder %q: %s", args[1], err)
			} else if isServer {
				exit.Errorf("%q does not look like a folder name", args[1])
			}

			if isServer, where, err := groupOrServer(args[0]); err != nil {
				exit.Errorf("failed to resolve restoration target %q: %s", args[0], err)
			} else if isServer {
				if reqID, err = client.RestoreServer(where, dest); err != nil {
					fmt.Fprintf(os.Stderr, "ERROR restoring server %s into %s: %s\n", where, args[1], err)
				} else {
					log.Printf("Restoring server %s: %s\n", args[0], reqID)
				}
			} else {
				if reqID, err = client.RestoreGroup(where, dest); err != nil {
					fmt.Fprintf(os.Stderr, "ERROR restoring group %s into %s %s\n", where, args[1], err)
				} else {
					log.Printf("Restoring group %s: %s\n", where, reqID)
				}
			}

			if reqID != "" {
				client.PollStatusFn(reqID, intvl, func(s clcv2.QueueStatus) {
					log.Printf("Restoring %s into %s: %s", args[0], args[1], s)
				})
			}
		},
	})
}
