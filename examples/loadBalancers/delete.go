/*
 * Delete a shared load balancer.
 */
package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
)

func main() {
	var location = flag.String("l", "", "Data centre location alias to use")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <Load Balancer ID>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 1 || *location == "" {
		flag.Usage()
		os.Exit(1)
	}

	client, err := clcv2.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	if err := client.DeleteSharedLoadBalancer(flag.Arg(0), *location); err != nil {
		exit.Fatalf("failed to delete load balancer: %s", err)
	}

	fmt.Printf("Successfully deleted load balancer %s.\n", flag.Arg(0))
}
