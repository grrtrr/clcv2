/*
 * Delete a given cross-datacenter firewall policy.
 */
package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/grrtrr/clcv2/clcv2cli"
	"github.com/grrtrr/exit"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <Location> <PolicyID>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(1)
	}

	client, err := clcv2cli.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	if err := client.DeleteCrossDataCenterFirewallPolicy(flag.Arg(0), flag.Arg(1)); err != nil {
		exit.Fatalf("failed to delete %s firewall policy %s: %s", flag.Arg(0), flag.Arg(1), err)
	}
	fmt.Printf("Successfully deleted cross-datacenter policy %s in %s.\n", flag.Arg(1), strings.ToUpper(flag.Arg(0)))
}
