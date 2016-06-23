/*
 * Remove one or more (non-system) disks from a server.
 */
package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"regexp"

	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
)

func main() {
	// Allow two types of ID: (a) <major>:<minor> syntax, (b) <minor> syntax
	var reMajMin = regexp.MustCompile(`^\d+:\d+$`)
	var reMin = regexp.MustCompile(`^\d+$`)
	var ids []string

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <Server Name> <diskId> [<diskId> ...]\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() < 2 {
		flag.Usage()
		os.Exit(1)
	}

	for i := 1; i < flag.NArg(); i++ {
		if reMajMin.MatchString(flag.Arg(i)) {
			ids = append(ids, flag.Arg(i))
		} else if reMin.MatchString(flag.Arg(i)) {
			ids = append(ids, fmt.Sprintf("0:%s", flag.Arg(i)))
		} else {
			exit.Errorf("invalid disk ID %q", flag.Arg(i))
		}
	}

	client, err := clcv2.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	server, err := client.GetServer(flag.Arg(0))
	if err != nil {
		exit.Fatalf("Failed to list details of server %q: %s", flag.Arg(0), err)
	}

	disks := make([]clcv2.ServerAdditionalDisk, 0)
	for i := range server.Details.Disks {
		if inStringArray(server.Details.Disks[i].Id, ids...) {
			fmt.Printf("Deleting disk %s (%d GB) ...\n", server.Details.Disks[i].Id, server.Details.Disks[i].SizeGB)
		} else {
			disks = append(disks, clcv2.ServerAdditionalDisk{
				Id:     server.Details.Disks[i].Id,
				SizeGB: server.Details.Disks[i].SizeGB,
			})
		}
	}

	statusId, err := client.ServerSetDisks(flag.Arg(0), disks)
	if err != nil {
		exit.Fatalf("Failed to update the disk configuration on %q: %s", flag.Arg(0), err)
	}

	fmt.Printf("Status Id for updating the disks on %s: %s\n", flag.Arg(0), statusId)
}

// inStringArray returns true if @s is found in @cand
func inStringArray(s string, cand ...string) bool {
	for _, c := range cand {
		if c == s {
			return true
		}
	}
	return false
}
