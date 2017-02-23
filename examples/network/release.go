/*
 * Release a given network in a given data centre.
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
	var location = flag.String("l", "", "Data centre alias of the network")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <network-ID>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() != 1 || *location == "" {
		flag.Usage()
		os.Exit(1)
	}

	client, err := clcv2cli.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	if err := client.ReleaseNetwork(*location, flag.Arg(0)); err != nil {
		exit.Fatalf("failed to release network %s in %s: %s", flag.Arg(0), *location, err)
	}
	fmt.Printf("Successfully released network %s in %s.\n", flag.Arg(0), strings.ToUpper(*location))
}
