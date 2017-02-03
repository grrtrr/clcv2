package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/grrtrr/exit"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var MkDir = &cobra.Command{
	Use:     "mkdir <new-folder>  [parent-folder]",
	Aliases: []string{"md"},
	Short:   "Create a new folder",
	Long:    "Create a new folder (hardware group). If no parent-folder is specified, create new folder at the root of current data centre (-l argument)",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if l := len(args); l < 1 || l > 2 {
			return errors.Errorf("Need at least a name for a folder to create (parent folder name is optional")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		var parentGroup string

		if l := len(args); l == 2 {
			if isServer, uuid, err := groupOrServer(args[1]); err != nil {
				exit.Fatalf("failed to look up parent group %s: %s", args[1], err)
			} else if isServer {
				exit.Errorf("%q does not look like a (parent) group name", args[1])
			} else {
				parentGroup = uuid
			}
		} else if conf.Location == "" {
			exit.Errorf("need a location argument (-l) if no parent folder is given")
		} else if root, err := client.GetGroups(conf.Location); err != nil {
			exit.Fatalf("ailed to look up group root folder in %s: %s", conf.Location, err)
		} else {
			fmt.Printf("Creating %q at at %s data centre root (%q)\n", args[0], strings.ToUpper(conf.Location), root.Name)
			parentGroup = root.Id
		}

		if g, err := client.CreateGroup(args[0], parentGroup, fmt.Sprintf("New subfolder %s", args[0]), nil); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create folder %s: %s\n", args[0], err)
		} else {
			fmt.Printf("New folder %q  with UUID %s\n", g.Name, g.Id)
		}
	},
}

func init() {
	Root.AddCommand(MkDir)
}
