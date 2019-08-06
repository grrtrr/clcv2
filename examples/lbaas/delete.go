/*
 * Delete a load balancer.
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
	var location = flag.String("l", "", "Data centre location alias to use")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <Load Balancer UUID>\n", path.Base(os.Args[0]))
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

	if err := client.DeleteLbInstance(flag.Arg(0), *location); err != nil {
		exit.Fatalf("failed to delete load balancer: %s", err)
	}

	fmt.Printf("Successfully deleted load balancer instance %s\n", flag.Arg(0))
}
