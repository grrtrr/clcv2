/*
 * Set the parent group of an existing HW group.
 */
package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/grrtrr/clcv2/clcv2cli"
	"github.com/grrtrr/exit"
)

func main() {
	var child string /* UUID of the group to relocate */
	var parent = flag.String("g", "", "UUID or name of the new parent group")
	var location = flag.String("l", "", "Location to use if using Group Name instead of UUID")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <Group Name or UUID>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() != 1 || *parent == "" {
		flag.Usage()
		os.Exit(1)
	}

	client, err := clcv2cli.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	if _, err := hex.DecodeString(flag.Arg(0)); err == nil {
		child = flag.Arg(0)
	} else if *location == "" {
		exit.Errorf("Need a location argument (-l) if not using Group UUID (%s)", flag.Arg(0))
	} else {
		if grp, err := client.GetGroupByName(flag.Arg(0), *location); err != nil {
			exit.Errorf("failed to resolve group name %q: %s", flag.Arg(0), err)
		} else if grp == nil {
			exit.Errorf("No group named %q was found in %s", flag.Arg(0), *location)
		} else {
			child = grp.Id
		}
	}

	if _, err := hex.DecodeString(*parent); err == nil {
		/* Looks like a Group UUID */
	} else if *location == "" {
		exit.Errorf("Need a location argument (-l) if parent (-g %s) is not a UUID", *parent)
	} else {
		if grp, err := client.GetGroupByName(*parent, *location); err != nil {
			exit.Errorf("failed to resolve group name %q: %s", *parent, err)
		} else if grp == nil {
			exit.Errorf("No group named %q was found in %s", *parent, *location)
		} else {
			*parent = grp.Id
		}
	}

	err = client.GroupSetParent(child, *parent)
	if err != nil {
		exit.Fatalf("failed to change the parent group of %q: %s", flag.Arg(0), err)
	}

	fmt.Printf("Successfully changed the parent group of %s to %s.\n", flag.Arg(0), *parent)
}
