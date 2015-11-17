/*
 * Power server off.
 *
 * This is a forced shutdown of a server. It's the equivalent to unplugging a physical machine.
 * All memory and CPU charges stop accruing, monitors are disabled, and the machine ends up in
 * a powered off state. Any licensing charges (if applicable) and storage charges continue accruing.
 * If the server is moved to archive storage, then any applicable licensing charges cease.
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

	client, err := clcv2.NewClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	statusId, err := client.PowerOffServer(flag.Arg(0))
	if err != nil {
		exit.Fatalf("Failed to power off server %s: %s", flag.Arg(0), err)
	}

	fmt.Println("Request ID for powering server off:", statusId)
}
