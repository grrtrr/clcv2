/*
 * Set the parent group of an existing server.
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
	var group = flag.String("g", "", "UUID or name of the new parent group")
	var location = flag.String("l", "", "Location to use if -g refers to a Group-Name")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <server-name>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() != 1 || *group == "" {
		flag.Usage()
		os.Exit(1)
	}

	client, err := clcv2cli.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	if _, err := hex.DecodeString(*group); err == nil {
		/* Looks like a Group UUID */
	} else if *location == "" {
		exit.Errorf("Need a location argument (-l) if -g (%s) is not a UUID", *group)
	} else {
		if grp, err := client.GetGroupByName(*group, *location); err != nil {
			exit.Errorf("failed to resolve group name %q: %s", *group, err)
		} else if grp == nil {
			exit.Errorf("No group named %q was found in %s", *group, *location)
		} else {
			*group = grp.Id
		}
	}

	err = client.ServerSetGroup(flag.Arg(0), *group)
	if err != nil {
		exit.Fatalf("failed to change the parent group on %q: %s", flag.Arg(0), err)
	}

	fmt.Printf("Successfully changed the parent group of %s to %s.\n", flag.Arg(0), *group)
}
