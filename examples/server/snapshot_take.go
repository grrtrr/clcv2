/*
 * Take a snapshot of a server.
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
	var days = flag.Int("days", 10, "Number of days to keep the snapshot for")

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

	statusId, err := client.SnapshotServer(flag.Arg(0), *days)
	if err != nil {
		exit.Fatalf("Failed to take snapshot of server %s: %s", flag.Arg(0), err)
	}

	fmt.Println("Request ID for taking server snapshot:", statusId)
}
