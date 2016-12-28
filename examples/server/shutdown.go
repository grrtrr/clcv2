/*
 * Initiates a graceful shutdown of the corresponding server.
 *
 * Like the "off" power command, all memory and CPU charges cease, monitors are disabled,
 * and the machine is left in a powered off state.
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

	statusId, err := client.ShutdownServer(flag.Arg(0))
	if err != nil {
		exit.Fatalf("failed to shut down server %s: %s", flag.Arg(0), err)
	}

	fmt.Println("Request ID for server shut-down:", statusId)
}
