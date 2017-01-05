package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	Root.AddCommand(&cobra.Command{
		Use:     "rename  <group>  <newName>",
		Short:   "Rename group",
		Long:    "Assign @newName to existing group @group",
		PreRunE: checkArgs(2, "Need name of existing group and the new name to use for it"),
		Run: func(cmd *cobra.Command, args []string) {
			if isServer, where, err := groupOrServer(args[0]); err != nil {
				fmt.Printf("Failed to look up group %s: %s\n", args[0], err)
			} else if isServer {
				fmt.Printf("ERROR: %q does not look like a group name\n", args[0])
			} else if err := client.GroupSetName(where, args[1]); err != nil {
				fmt.Printf("ERROR: failed to rename %s: %s\n", args[0], err)
			} else {
				fmt.Printf("Successfully renamed %s into %s\n", args[0], args[1])
			}
		},
	})
}
