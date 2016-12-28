/*
 * List the snapshot of a server.
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

	snapshot, err := client.GetServerSnapshot(flag.Arg(0))
	if err != nil {
		exit.Fatalf("failed to query snapshots of %s: %s", flag.Arg(0), err)
	}

	if snapshot == nil {
		fmt.Printf("Server %s does not have any snapshots.\n", flag.Arg(0))
	} else {
		fmt.Printf("Snapshot of %s: %s\n", flag.Arg(0), snapshot.Name)
	}
}
