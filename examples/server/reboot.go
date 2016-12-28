/*
 * Reboot server.
 *
 * Executes a graceful reboot of the target server. Unlike the forced "reset" power
 * command, this instructs the operating system to initiate a proper stop and restart.
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

	statusId, err := client.RebootServer(flag.Arg(0))
	if err != nil {
		exit.Fatalf("failed to reboot server %s: %s", flag.Arg(0), err)
	}

	fmt.Println("Request ID for server reboot:", statusId)
}
