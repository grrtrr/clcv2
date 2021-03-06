/*
 * Reset server (forced power-cycle).
 *
 * Similar to the relationship between "Off" and "Shut Down", the reset command is
 * a forced power-off and power-on combination.
 * It is equivalent to the reset button on a physical computer.
 */
package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/grrtrr/clcv2/clcv2cli"
	"github.com/grrtrr/exit"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <server-name>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	client, err := clcv2cli.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	statusId, err := client.ResetServer(flag.Arg(0))
	if err != nil {
		exit.Fatalf("failed to reset server %s: %s", flag.Arg(0), err)
	}

	fmt.Println("Request ID for server reset:", statusId)
}
