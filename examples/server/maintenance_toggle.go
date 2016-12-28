/*
 * Enable or disable maintenance mode on a server
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
	var maintenance = flag.Bool("m", true, "Turn maintenance mode on (-m) or off (-m=false)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <server-name>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
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

	statusId, err := client.ServerSetMaintenance(flag.Arg(0), *maintenance)
	if err != nil {
		exit.Fatalf("failed to modify maintenance mode on server %s: %s", flag.Arg(0), err)
	}

	fmt.Println("Request ID for maintenance mode status change:", statusId)
}
