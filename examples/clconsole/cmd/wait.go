package cmd

import (
	"log"

	"github.com/grrtrr/clcv2"
	"github.com/spf13/cobra"
)

func init() {
	Root.AddCommand(&cobra.Command{
		Use:     "wait  <statusID>",
		Aliases: []string{"job", "status"},
		Short:   "Await completion of queue job and report status",
		PreRunE: checkArgs(1, "Need a status ID to poll"),
		Run: func(cmd *cobra.Command, args []string) {
			client.PollStatusFn(args[0], intvl, func(s clcv2.QueueStatus) {
				log.Printf("%s: %s", args[0], s)
			})
		},
	})
}
