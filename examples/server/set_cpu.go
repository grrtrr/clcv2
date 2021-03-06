/*
 * Change the amount of CPU cores on an existing server.
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
	var cpus = flag.Int("cpu", 0, "The number of CPU cores to set for this server")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <server-name>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() != 1 || *cpus == 0 {
		flag.Usage()
		os.Exit(1)
	}

	client, err := clcv2cli.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	statusId, err := client.ServerSetCpus(flag.Arg(0), fmt.Sprint(*cpus))
	if err != nil {
		exit.Fatalf("failed to change the number of CPUs on %q: %s", flag.Arg(0), err)
	}

	fmt.Printf("Status Id for changing the #CPUs on %s: %s\n", flag.Arg(0), statusId)
}
