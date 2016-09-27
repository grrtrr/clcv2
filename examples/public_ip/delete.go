/*
 * Release the given public IP address of a server so that it is no longer associated with the server
 * and available to be claimed again by another server.
 */
package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <server-name> <public-IP>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(1)
	}

	client, err := clcv2.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	statusId, err := client.RemovePublicIPAddress(flag.Arg(0), flag.Arg(1))
	if err != nil {
		exit.Fatalf("failed to remove public IP on %q: %s", flag.Arg(0), err)
	}
	client.PollStatus(statusId, 1*time.Second)
}
