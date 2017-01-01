package main

import (
	"fmt"
	"os"

	"github.com/grrtrr/clcv2/examples/clconsole/cmd"
	"github.com/spf13/cobra"
)

var ()

/*
 * Initialization
 */
func init() {
	cobra.EnableCommandSorting = false

}

/*
 * Main entry point
 */
func main() {
	if err := cmd.Root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
}
