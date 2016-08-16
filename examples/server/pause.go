/*
 * Pause a server.
 *
 * When a virtual machine is paused, its state is frozen (e.g. memory, open applications)
 * and monitoring ceases. Billing charges for CPU and memory stop. A paused machine can be
 * quickly brought back to life by issuing the "On" power command.
 * Any applicable licensing charges continue to accrue while a machine is paused.
 */
package main

import (
	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
	"path"
	"flag"
	"fmt"
	"os"
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

	client, err := clcv2.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	statusId, err := client.PauseServer(flag.Arg(0))
	if err != nil {
		exit.Fatalf("failed to pause server %s: %s", flag.Arg(0), err)
	}

	fmt.Printf("Request ID for pausing server: %s\n", statusId)
}
