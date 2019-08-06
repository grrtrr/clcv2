/*
 * Poll 'Create LBaaS' request'
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
		intvl    = flag.Duration("i", 5*time.Second, "Poll interval (use 0 to disable polling)")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <Request ID>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()

	/* The Location argument is always required */
	if flag.NArg() != 1 || *location == "" {
		flag.Usage()
		os.Exit(1)
	}

	client, err := clcv2cli.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	for {
		req, err := client.GetLbCreateRequest(*location, flag.Arg(0))
		if err != nil {
			exit.Fatalf("failed to query LBaaS request %s status: %s", flag.Arg(0), err)
		}
		fmt.Printf("Instance: %s (%s)", req.Links[0].ResourceId, req.Status)
		fmt.Printf(", created %s", humanize.Time(req.Created.Time()))
		if req.Completed != nil {
			fmt.Printf(", completed %s, runtime %s", humanize.Time(req.Completed.Time()),
				req.Completed.Time().Sub(req.Created.Time()).Round(10*time.Millisecond))
		}
		fmt.Println("")
		if *intvl == 0 || req.Completed != nil {
			break
		}
		time.Sleep(*intvl)
	}
}
