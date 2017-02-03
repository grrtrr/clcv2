// Cobra commandline console driver
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/grrtrr/clcv2/examples/clconsole/cmd"
	"github.com/spf13/cobra"
)

func main() {
	// Logging format - we don't need date
	log.SetFlags(log.Ltime | log.Lshortfile)

	// Do not sort the commands alphabetically
	cobra.EnableCommandSorting = false

	defer cmd.ExitHandler()

	if err := cmd.Root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
}
