package cmd

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/grrtrr/clcv2"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

/*
 * Helper Functions
 */

// die is like die in Perl
func die(format string, a ...interface{}) {
	format = fmt.Sprintf("%s: %s\n", path.Base(os.Args[0]), strings.TrimSpace(format))
	fmt.Fprintf(os.Stderr, format, a...)
	os.Exit(1)
}

// checkArgs returns a cobra-compatible PreRunE argument-validation function
func checkArgs(nargs int, errMsg string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) != nargs {
			return fmt.Errorf(errMsg)
		}
		return nil
	}
}

// extractServerNames extracts all server names specified via @args
// - server names contained in @args are returned directly
// - for group names, it recursively collects all server names contained in the group
func extractServerNames(args []string) (ret []string, err error) {
	var root *clcv2.Group
	var groupCallback = func(g *clcv2.Group, arg interface{}) interface{} {
		for _, l := range g.Links {
			if l.Rel == "server" {
				ret = append(ret, l.Id)
			}
		}
		return nil
	}

	for _, name := range args {
		if isServer, where, err := groupOrServer(name); isServer {
			ret = append(ret, where)
		} else if location == "" {
			return nil, errors.Errorf("Location argument (-l) is required in order to traverse group %s", name)
		} else if root, err = client.GetGroups(location); err != nil {
			return nil, errors.Errorf("Failed to look up groups at %s: %s\n", location, err)
		} else {
			start := root
			if where != "" {
				start = clcv2.FindGroupNode(root, func(g *clcv2.Group) bool { return g.Id == where })
				if start == nil {
					return nil, errors.Errorf("Failed to look up UUID %s in %s - is this the correct value?", where, location)
				}
			}
			clcv2.VisitGroupHierarchy(start, groupCallback, nil)
		}
	}
	return ret, nil
}
