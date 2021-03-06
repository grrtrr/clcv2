/*
 * Gets the status of a particular job in the queue,
 * which keeps track of any long-running asynchronous requests
 * (such as server power operations or provisioning tasks).
 */
package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/grrtrr/clcv2/clcv2cli"
	"github.com/grrtrr/exit"
)

func main() {
	var intvl = flag.Duration("i", 5*time.Second, "Poll interval (use 0 to disable polling)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <RequestID>\n", path.Base(os.Args[0]))
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

	client.PollStatus(flag.Arg(0), *intvl)
}
