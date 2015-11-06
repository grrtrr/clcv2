/*
 * Set the description of an existing server.
 */
package main

import (
	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
	"path"
	"flag"
	"log"
	"fmt"
	"os"
)

func main() {
	var description = flag.String("desc", "", "New description to use for the server")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <server-name>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() != 1 || *description == "" {
		flag.Usage()
		os.Exit(1)
	}

	client, err := clcv2.NewClient(nil, log.New(os.Stdout, "", log.LstdFlags | log.Ltime))
	if err != nil {
		exit.Fatal(err.Error())
	}

	err = client.ServerSetDescription(flag.Arg(0), *description)
	if err != nil {
		exit.Fatalf("Failed to change the description on %q: %s", flag.Arg(0), err)
	}

	fmt.Printf("Successfully changed the description on %s.\n", flag.Arg(0))
}
