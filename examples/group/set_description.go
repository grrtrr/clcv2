/*
 * Set the description of an existing HW group.
 */
package main

import (
	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
	"encoding/hex"
	"path"
	"flag"
	"fmt"
	"os"
)

func main() {
	var group string	/* UUID of the group to change */
	var newDesc  = flag.String("desc", "", "New description to use for the group")
	var location = flag.String("l", "", "Location to use if using a Group-Name instead of a UUID")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <Group Name or UUID>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() != 1 || *newDesc == "" {
		flag.Usage()
		os.Exit(1)
	}

	client, err := clcv2.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	if _, err := hex.DecodeString(flag.Arg(0)); err == nil {
		group = flag.Arg(0)
	} else if  *location == "" {
		exit.Errorf("Need a location argument (-l) if not using Group UUID (%s)", flag.Arg(0))
	} else {
		if grp, err := client.GetGroupByName(flag.Arg(0), *location); err != nil {
			exit.Errorf("Failed to resolve group name %q: %s", flag.Arg(0), err)
		} else if grp == nil {
			exit.Errorf("No group named %q was found in %s", flag.Arg(0), *location)
		} else {
			group = grp.Id
		}
	}

	if err = client.GroupSetDescription(group, *newDesc); err != nil {
		exit.Fatalf("Failed to change the description of %q: %s", flag.Arg(0), err)
	}

	fmt.Printf("Successfully changed the description of %s to %q.\n", flag.Arg(0), *newDesc)
}
