package cmd

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/grrtrr/clcv2"
	"github.com/spf13/cobra"
)

var On = &cobra.Command{
	Use:     "on [group|server [group|server]...]",
	Aliases: []string{"start", "up"},
	Short:   "Power on server(s) (or resume from paused state)",
	Run: func(cmd *cobra.Command, args []string) {
		var wg sync.WaitGroup

		servers, err := extractServerNames(args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
			return
		}
		for _, name := range servers {
			name := name
			wg.Add(1)
			go func() {
				defer wg.Done()

				if reqID, err := client.PowerOnServer(name); err != nil {
					fmt.Fprintf(os.Stderr, "ERROR powering on server %s: %s\n", name, err)
				} else {
					log.Printf("%s power-on: %s", name, reqID)

					client.PollStatusFn(reqID, intvl, func(s clcv2.QueueStatus) {
						log.Printf("%s power-on: %s", name, s)
					})
				}
			}()
		}
		wg.Wait()
	},
}

func init() {
	Root.AddCommand(On)
}
