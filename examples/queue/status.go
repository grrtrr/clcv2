/*
 * Gets the status of a particular job in the queue,
 * which keeps track of any long-running asynchronous requests
 * (such as server power operations or provisioning tasks).
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

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <RequestID>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	client, err := clcv2.NewClient(nil, log.New(os.Stdout, "", log.LstdFlags | log.Ltime))
	if err != nil {
		exit.Fatal(err.Error())
	}

	status, err := client.GetStatus(flag.Arg(0))
	if err != nil {
		exit.Fatalf("Failed to list queue requests: %s", err)
	}

	fmt.Printf("Status: %s\n", status)
	fmt.Println(status == clcv2.Succeeded)
}
