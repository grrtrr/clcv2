/*
 * Delete a server by name
 */
package main

import (
	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
	"github.com/kr/pretty"
	"path"
	"flag"
	"log"
	"fmt"
	"os"
)

func main() {

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <server-name>\n", path.Base(os.Args[0]))
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

	status, err := client.DeleteServer(flag.Arg(0))
	if err != nil {
		exit.Fatalf("Failed to delete server %s: %s", flag.Arg(0), err)
	}

	fmt.Printf("Status of %s: ", status.Server)
	if status.IsQueued {
		fmt.Printf("queued\n")
	} else {
		fmt.Printf("not queued\n")
	}

	if status.ErrorMessage != "" {
		fmt.Printf("ERROR: %s\n", status.ErrorMessage)
	}

	fmt.Println("Links:")
	pretty.Println(status.Links)
}
