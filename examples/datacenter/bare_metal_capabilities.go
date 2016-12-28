/*
 * For a given (or default) datacenter, list its bare-metal capabilities.
 */
package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/grrtrr/clcv2/clcv2cli"
	"github.com/grrtrr/exit"
	"github.com/kr/pretty"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  Location-Alias\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
		os.Exit(0)
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

	capa, err := client.GetBareMetalCapabilities(flag.Arg(0))
	if err != nil {
		exit.Fatalf("failed to query bare-metal capabilities of %s: %s", flag.Arg(0), err)
	}

	fmt.Printf("Datacenter %s:\n", flag.Arg(0))
	pretty.Println(capa)
}
