/*
 * List the snapshot of a server.
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

	snapshot, err := client.GetServerSnapshot(flag.Arg(0))
	if err != nil {
		exit.Fatalf("Failed to query snapshots of %s: %s", flag.Arg(0), err)
	}

	if snapshot == nil {
		fmt.Printf("Server %s does not have any snapshots.\n", flag.Arg(0))
	} else {
		fmt.Printf("Snapshot of %s: %s\n", flag.Arg(0), snapshot.Name)
	}
}
