/*
 * Release the given public IP address of a server so that it is no longer associated with the server
 * and available to be claimed again by another server.
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
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <server-name> <public-IP>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(1)
	}

	client, err := clcv2.NewClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	reqId, err := client.RemovePublicIPAddress(flag.Arg(0), flag.Arg(1))	if err != nil {
		exit.Fatalf("Failed to remove public IP on %q: %s", flag.Arg(0), err)
	}

	fmt.Println("Request ID for removing public IP:", reqId)
}
