/*
 * Revert to a snapshot of a server.
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
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <Server-name>\n", path.Base(os.Args[0]))
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

	snap, statusId, err := client.RevertToSnapshot(flag.Arg(0))
	if err != nil {
		exit.Fatalf("Failed to revert %s to snapshot: %s", flag.Arg(0), err)
	}

	fmt.Printf("Request ID for reverting %s to snapshot %q: %s\n", flag.Arg(0), snap.Name, statusId)
}
