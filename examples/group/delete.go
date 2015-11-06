/*
 * Delete a Hardware Group via its UUID
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
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <HW Group UUID>\n", path.Base(os.Args[0]))
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

	reqId, err := client.DeleteGroup(flag.Arg(0))
	if err != nil {
		exit.Fatalf("Failed to delete hardware group: %s", err)
	}

	fmt.Printf("Status ID for %s deletion request: %s\n", flag.Arg(0), reqId)
}
