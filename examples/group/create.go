/*
 * Create a new hardware group
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
	var parentGroup = flag.String("g", "", "UUID or Name (if unique and -l present) of the parent Hardware Group")
	var location    = flag.String("l", "", "Data centre location to use for resolving -g <Group-Name>")
	var desc        = flag.String("t", "", "Textual description of the new group")
	var parentUUID string

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <New Group Name>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 1 || *parentGroup == "" {
		flag.Usage()
		os.Exit(1)
	}

	/* parentGroup may be hex uuid or group name */
	if _, err := hex.DecodeString(*parentGroup); err == nil {
		parentUUID = *parentGroup
	} else if *location == "" {
		exit.Errorf("Using -g <Group-Name> requires -l <Location> to be set")
	}

	client, err := clcv2.NewClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	if parentUUID == "" {
		var group *clcv2.Group

		if group, err = client.GetGroupByName(*parentGroup, *location); err != nil {
			exit.Errorf("Failed to resolve group name %q: %s", *parentGroup, err)
		} else if group == nil {
			exit.Errorf("No group named %q was found in %s", *parentGroup, *location)
		} else {
			parentUUID = group.Id
		}
	}

	g, err := client.CreateGroup(flag.Arg(0), parentUUID, *desc, []clcv2.SimpleCustomField{})
	if err != nil {
		exit.Fatalf("Failed to create hardware group %q: %s", flag.Arg(0), err)
	}

	fmt.Println("New Group: ", g.Name)
	fmt.Println("UUID:      ", g.Id)
}
