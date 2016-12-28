/*
 * Power server on (or resume from a paused state).
 *
 * Initiates the operating system boot sequence. Billing charges for memory,
 * CPU, and licenses (if applicable) start accruing, and monitors are re-enabled.
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

	statusId, err := client.PowerOnServer(flag.Arg(0))
	if err != nil {
		exit.Fatalf("failed to power server %s on: %s", flag.Arg(0), err)
	}

	fmt.Println("Request ID for powering on server:", statusId)
}
