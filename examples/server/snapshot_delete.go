/*
 * Delete a snapshot for a specified server
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
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <Server-Name>\n", path.Base(os.Args[0]))
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

	sn, statusId, err := client.DeleteSnapshot(flag.Arg(0))
	if err != nil {
		exit.Fatalf("Failed to delete snapshot on %s: %s", flag.Arg(0), err)
	}

	fmt.Printf("Request ID for deleting snapshot %q: %s\n", sn.Name, statusId)
}