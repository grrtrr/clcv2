package cmd

import (
	"flag"
	"fmt"

	"github.com/grrtrr/exit"
	"github.com/spf13/cobra"
)

var mv = &cobra.Command{
	Use:     "mv  <server|group>  <destination>",
	Short:   "Move group/server to different folder",
	PreRunE: checkArgs(2, "Need folder names for source server/group and destination group"),
	Run: func(cmd *cobra.Command, args []string) {
		if isServer, newParent, err := groupOrServer(args[1]); err != nil {
			fmt.Printf("Failed to look up parent group %s: %s\n", args[1], err)
		} else if isServer {
			fmt.Printf("Destination %q does not look like a folder name\n", args[1])
		} else if isServer, where, err := groupOrServer(args[0]); err != nil {
			fmt.Printf("Error looking up %s: %s\n", args[0], err)
		} else if isServer { // moving a server
			if err = client.ServerSetGroup(where, newParent); err != nil {
				exit.Fatalf("failed to change the parent group on %q: %s", where, err)
			} else {
				fmt.Printf("Successfully changed the parent group of %s to %s.\n", where, args[1])
			}
		} else { // moving a folder
			if where == newParent {
				fmt.Printf("Nothing to do - source is the same as destination.\n")
			} else if err = client.GroupSetParent(where, newParent); err != nil {
				exit.Fatalf("failed to change the parent group on %q: %s", where, err)
			} else {
				fmt.Printf("Successfully changed the parent group of %s to %s.\n", where, flag.Arg(2))
			}
		}
	},
}

func init() {
	Root.AddCommand(mv)
}
