package cmd

import (
	"fmt"
	"os"
	"path"
	"strings"

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
