// Cobra commandline console driver
package main

import (
	"fmt"
	"os"

	"github.com/grrtrr/clcv2/examples/clconsole/cmd"
)

func main() {
	if err := cmd.Root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
}
