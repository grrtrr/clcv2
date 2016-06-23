/*
 * Archive a hardware group.
 */
package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
)

func main() {
	var location = flag.String("l", "", "Location to use if using a Group-Name instead of a UUID")
	var uuid string

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <Group Name or UUID>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	client, err := clcv2.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	if _, err := hex.DecodeString(flag.Arg(0)); err == nil {
		uuid = flag.Arg(0)
	} else if *location == "" {
		exit.Errorf("Need a location argument (-l) if not using Group UUID (%s)", flag.Arg(0))
	} else {
		if grp, err := client.GetGroupByName(flag.Arg(0), *location); err != nil {
			exit.Errorf("Failed to resolve group name %q: %s", flag.Arg(0), err)
		} else if grp == nil {
			exit.Errorf("No group named %q was found in %s", flag.Arg(0), *location)
		} else {
			uuid = grp.Id
		}
	}

	statusId, err := client.ArchiveGroup(uuid)
	if err != nil {
		exit.Fatalf("Failed to archive group %s: %s", flag.Arg(0), err)
	}

	fmt.Println("Request ID for archiving group:", statusId)
}
