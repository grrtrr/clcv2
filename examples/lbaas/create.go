/*
 * Create a new load balancer.
 */
package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/grrtrr/clcv2/clcv2cli"
	"github.com/grrtrr/exit"
)

func main() {
	var (
		location = flag.String("l", "", "Location alias of data centre to host load balancer")
		desc     = flag.String("desc", "", "Textual description of the load balancer")
		intvl    = flag.Duration("i", 5*time.Second, "Poll interval (use 0 to disable polling)")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <Load Balancer Name>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	fmt.Println(flag.NArg())
	if flag.NArg() != 1 || *location == "" {
		flag.Usage()
		os.Exit(1)
	}

	client, err := clcv2cli.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	req, err := client.CreateLbInstance(flag.Arg(0), *desc, *location)
	if err != nil {
		exit.Fatalf("failed to create load balancer %q: %s", flag.Arg(0), err)
	}

	fmt.Printf("Creating %s load balancer %s (%s)\n", *location, req.Links[0].ResourceId, req.Status)

	for *intvl > 0 && req.Completed == nil {
		time.Sleep(*intvl)

		req, err = client.GetLbCreateRequest(*location, req.ID.String())
		if err != nil {
			exit.Fatalf("failed to query LBaaS request %s status: %s", req.ID, err)
		}
		fmt.Printf("Instance: %s (%s)", req.Links[0].ResourceId, req.Status)
		fmt.Printf(", created %s", humanize.Time(req.Created.Time()))
		if req.Completed != nil {
			fmt.Printf(", completed %s, runtime %s", humanize.Time(req.Completed.Time()),
				req.Completed.Time().Sub(req.Created.Time()).Round(10*time.Millisecond))
		}
		fmt.Println("")
	}
}
