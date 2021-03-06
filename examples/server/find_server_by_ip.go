/*
 * Identify a server given only one of its IP addresses
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
	var location = flag.String("l", "", "Alias of the data centre the server resides in")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <IP Address>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	} else if *location == "" {
		exit.Errorf("need a location argument (-l)")
	}

	client, err := clcv2cli.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	iad, err := client.GetNetworkDetailsByIp(flag.Arg(0), *location)
	if err != nil {
		exit.Fatalf("failed to look up %s: %s", flag.Arg(0), err)
	} else if iad == nil {
		exit.Errorf("No match found for %s in %s", flag.Arg(0), *location)
	}

	// The 'Server' field is not necessarily filled in, hence we need to test here.
	if iad.Server != "" {
		fmt.Printf("%s is used by %s.\n", iad.Address, iad.Server)
	} else {
		fmt.Printf("%s is in %s use in %s, but the server name is not disclosed.\n", iad.Address, iad.Type, *location)
	}
}
