/*
 * Claim a new network in a given data centre.
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
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <data-centre>\n", path.Base(os.Args[0]))
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

	uri, err := client.ClaimNetwork(flag.Arg(0))
	if err != nil {
		exit.Fatalf("failed to claim a new network in %q: %s", flag.Arg(0), err)
	}
	s, err := client.PollClaimNetworkStatus(uri)
	fmt.Println("status", s, err)
	fmt.Printf("Successfully claimed a new network in %s.\n", flag.Arg(0))
}
