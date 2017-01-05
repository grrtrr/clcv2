package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var MkDir = &cobra.Command{
	Use:     "mkdir <newFolder>  <parentGroup>",
	Short:   "Create a new folder",
	Long:    "Create a new folder @newFolder (hardware group) inside @parentGroup",
	PreRunE: checkArgs(2, "Need folder names for new group and parent group"),
	Run: func(cmd *cobra.Command, args []string) {
		if isServer, where, err := groupOrServer(args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to look up parent group %s: %s\n", args[1], err)
		} else if isServer {
			fmt.Fprintf(os.Stderr, "Does not look like a group name - %q\n", args[1])
		} else if g, err := client.CreateGroup(args[0], where, fmt.Sprintf("New subfolder %s", args[0]), nil); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create %s/%s: %s\n", args[1], args[0], err)
		} else {
			fmt.Printf("New subfolder of %s: %q (UUID: %s)\n", args[1], g.Name, g.Id)
		}
	},
}

func init() {
	Root.AddCommand(MkDir)
}
