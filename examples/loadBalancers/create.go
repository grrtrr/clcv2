/*
 * Create a new shared load balancer
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
	var location = flag.String("l", "", "Location alias of data centre to host load balancer")
	var desc = flag.String("desc", "", "Textual description of the load balancer")
	var status = flag.String("status", "enabled", "Whether to an 'enabled' or 'disabled' load balancer")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <Load Balancer Name>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	fmt.Println(flag.NArg())
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	client, err := clcv2.NewClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	lb, err := client.CreateSharedLoadBalancer(flag.Arg(0), *desc, *status, *location)
	if err != nil {
		exit.Fatalf("Failed to create shared load balancer %q: %s", flag.Arg(0), err)
	}

	fmt.Printf("Created %s load balancer %s, %q with IP %s\n", *location, lb.ID, lb.Name, lb.IpAddress)
}
