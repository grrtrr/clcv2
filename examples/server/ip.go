/*
 * Prints all IP addresses (public and private) associated with a given server.
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
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <Server Name> [<Server Name> ...]\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	client, err := clcv2cli.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	for _, srv := range flag.Args() {
		ips, err := client.GetServerIPs(srv)
		if err != nil {
			exit.Fatalf("failed to list details of server %q: %s", srv, err)
		}

		fmt.Printf("%-20s %s\n", srv+":", strings.Join(ips, ", "))
	}
}
