/*
 * Change the amount of memory on an existing server.
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
	var memory = flag.Int("mem", 0, "The amount of memory (in GB) to set for this server")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <server-name>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() != 1 || *memory == 0 {
		flag.Usage()
		os.Exit(1)
	}

	client, err := clcv2cli.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	statusId, err := client.ServerSetMemory(flag.Arg(0), fmt.Sprint(*memory))
	if err != nil {
		exit.Fatalf("failed to change the amount of Memory on %q: %s", flag.Arg(0), err)
	}

	fmt.Printf("Status Id for changing the memory on %s: %s\n", flag.Arg(0), statusId)
}
