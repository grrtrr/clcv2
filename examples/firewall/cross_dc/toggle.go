/*
 * Disable/enable a given cross-datacenter firewall policy.
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
	var enable = flag.Bool("enable", false, "Enable the policy")

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

	if err := client.ToggleCrossDataCenterFirewallPolicy(flag.Arg(0), flag.Arg(1), *enable); err != nil {
		exit.Fatalf("failed to change %s firewall policy %s status: %s", flag.Arg(0), flag.Arg(1), err)
	}
	fmt.Printf("Successfully set cross-datacenter policy %s enabled=%t.\n", flag.Arg(1), *enable)
}
