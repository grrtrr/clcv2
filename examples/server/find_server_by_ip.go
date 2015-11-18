/*
 * Identify a server given only one of its IP addresses
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
	var location = flag.String("l", "",  "Alias of the data centre the server resides in")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <IP Address>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() != 1 || *location == "" {
		flag.Usage()
		os.Exit(1)
	}

	client, err := clcv2.NewClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	iad, err := client.GetNetworkDetailsByIp(flag.Arg(0), *location)
	if err != nil {
		exit.Fatalf("Failed to lookup %s: %s", flag.Arg(0), err)
	} else if iad == nil {
		exit.Errorf("No match found for %s in %s", flag.Arg(0), *location)
	}

	fmt.Printf("%s is used by %s\n", iad.Address, iad.Server)
}
