/*
 * Delete a Hardware Group.
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
	var uuid string /* UUID of the HW group to delete */
	var location = flag.String("l", "", "Data center location if using HW Group-Name")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <HW Group-Name or UUID>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	client, err := clcv2.NewClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	/* If the first argument decodes as a hex value, assume it is a Hardware Group UUID */
	if _, err := hex.DecodeString(flag.Arg(0)); err == nil {
		uuid = flag.Arg(0)
	} else if *location == "" {
		exit.Errorf("Need a location (-l argument) when not using a HW Group UUID")
	} else {
		fmt.Printf("Resolving group UUID of %s in %s ...\n", flag.Arg(0), *location)
		if grp, err := client.GetGroupByName(flag.Arg(0), *location); err != nil {
			exit.Errorf("Failed to resolve group name %q: %s", flag.Arg(0), err)
		} else if grp == nil {
			exit.Errorf("No group named %q was found in %s", flag.Arg(0), *location)
		} else {
			uuid = grp.Id
		}
	}

	reqId, err := client.DeleteGroup(uuid)
	if err != nil {
		exit.Fatalf("Failed to delete hardware group: %s", err)
	}

	fmt.Printf("Status ID for group deletion: %s\n", reqId)
}
