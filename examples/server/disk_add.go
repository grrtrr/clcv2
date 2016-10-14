/*
 * Add a disk to a server.
 */
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
)

func main() {
	var size = flag.Int("size", 0, "Size of the disk in GB")
	var mount = flag.String("path", "", "Optional mountpoint (otherwise disk type defaults to 'raw')")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <Server Name>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() != 1 || *size <= 0 {
		flag.Usage()
		os.Exit(1)
	}

	client, err := clcv2.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	/* First get the list of disks */
	server, err := client.GetServer(flag.Arg(0))
	if err != nil {
		exit.Fatalf("failed to list details of server %q: %s", flag.Arg(0), err)
	}

	disks := make([]clcv2.ServerAdditionalDisk, len(server.Details.Disks))
	for i := range server.Details.Disks {
		disks[i] = clcv2.ServerAdditionalDisk{
			Id:     server.Details.Disks[i].Id,
			SizeGB: server.Details.Disks[i].SizeGB,
		}
	}

	/* Add disk at end - if path is specified, type must be set to 'partitioned'. */
	newDisk := clcv2.ServerAdditionalDisk{SizeGB: uint32(*size), Type: "raw"}
	if *mount != "" {
		newDisk.Path = *mount
		newDisk.Type = "partitioned"
	}

	reqID, err := client.ServerSetDisks(flag.Arg(0), append(disks, newDisk))
	if err != nil {
		exit.Fatalf("failed to update the disk configuration on %q: %s", flag.Arg(0), err)
	}

	log.Printf("Status Id for adding disk to %s: %s", flag.Arg(0), reqID)
	client.PollStatus(reqID, 5*time.Second)
}
