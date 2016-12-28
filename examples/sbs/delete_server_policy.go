/*
 * Delete a Server Policy given its policy ID.
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
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <Server-Policy-ID>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(0)
	}

	client, err := clcv2cli.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	if err := client.SBSdeleteServerPolicy(flag.Arg(0)); err != nil {
		exit.Fatalf("failed to delete SBS Server Policy %s: %s.", flag.Arg(0), err)
	}

	fmt.Printf("Successfully deleted Server Policy %s.\n", flag.Arg(0))
}
