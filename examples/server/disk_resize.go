/*
 * Resize an existing disk.
 */
package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"regexp"
	"time"

	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
)

func main() {
	var size = flag.Uint("size", 0, "New size of the disk in GB")
	// Allow the same ID types as in disk_remove.go
	var reMajMin = regexp.MustCompile(`^\d+:\d+$`)
	var reMin = regexp.MustCompile(`^\d+$`)
	var id string

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <Server-Name>  <Disk-ID>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() != 2 || *size == 0 {
		flag.Usage()
		os.Exit(1)
	} else if reMajMin.MatchString(flag.Arg(1)) {
		id = flag.Arg(1)
	} else if reMin.MatchString(flag.Arg(1)) {
		id = fmt.Sprintf("0:%s", flag.Arg(1))
	} else {
		exit.Errorf("invalid disk ID %q", flag.Arg(1))
	}

	client, err := clcv2.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

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
		if disks[i].Id == id {
			if disks[i].SizeGB == uint32(*size) {
				fmt.Printf("Disk %s size is already at %s.\n", id, *size)
				os.Exit(0)
			}
			fmt.Printf("Changing disk %s size from %d to %d GB) ...\n",
				id, disks[i].SizeGB, *size)
			disks[i].SizeGB = uint32(*size)
		}
	}

	statusId, err := client.ServerSetDisks(flag.Arg(0), disks)
	if err != nil {
		exit.Fatalf("failed to update the disk configuration on %q: %s", flag.Arg(0), err)
	}

	fmt.Printf("Status Id for adding disk to %s: %s\n", flag.Arg(0), statusId)
	client.PollStatus(flag.Arg(0), 10*time.Second)
}
