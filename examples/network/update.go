/*
 * Set the description of an existing server.
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
	var location = flag.String("l", "",    "Data centre alias of the network")
	var desc     = flag.String("desc", "", "Description of the VLAN")
	var name     = flag.String("name", "", "User-defined name of the network")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <Network-ID>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() != 1 || *location == "" || *name == "" {
		flag.Usage()
		os.Exit(1)
	}

	client, err := clcv2.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	err = client.UpdateNetwork(*location, flag.Arg(0), *name, *desc)
	if err != nil {
		exit.Fatalf("failed to update the name/description on %q: %s", flag.Arg(0), err)
	}

	fmt.Printf("Successfully changed the name/description of network %s.\n", flag.Arg(0))
}
