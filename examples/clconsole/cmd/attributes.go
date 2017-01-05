package cmd

import (
	"fmt"
	"log"
	"strconv"

	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	Root.AddCommand(&cobra.Command{
		Use:     "memory  <server> <memoryGB>",
		Aliases: []string{"mem"},
		Short:   "Set server memory",
		Long:    "Sets the memory of @server to size @memoryGB",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return errors.Errorf("Need a <server> and a <memoryGB> argument")
			} else if _, err := strconv.ParseUint(args[1], 10, 32); err != nil {
				return errors.Errorf("Invalid memoryGB value %q", args[1])
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Setting %s memory to %s GB ...\n", args[0], args[1])
			if reqID, err := client.ServerSetMemory(args[0], args[1]); err != nil {
				exit.Fatalf("failed to change the amount of Memory on %q: %s", args[0], err)
			} else {
				log.Printf("%s changing memory size: %s", args[0], reqID)

				client.PollStatusFn(reqID, intvl, func(s clcv2.QueueStatus) {
					log.Printf("%s changing memory size: %s", args[0], s)
				})
			}
		},
	})

	Root.AddCommand(&cobra.Command{
		Use:     "desc  <server>",
		Aliases: []string{"description"},
		Short:   "Power-off server(s)",
		Long:    "Do a forceful power-off of server(s) (as opposed to a soft OS-level shutdown)",
		PreRunE: checkArgs(2, "Need a server name and a new description for the server"),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Setting %s description to to %q.\n", args[0], args[1])

			if err := client.ServerSetDescription(args[0], args[1]); err != nil {
				exit.Fatalf("failed to change the description of %q: %s", args[0], err)
			}
		},
	})
}
