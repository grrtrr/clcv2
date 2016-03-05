/*
 * Restore a server from archive.
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
	var hwGroup = flag.String("g", "", "UUID or name (if unique) of the HW group to add this server to")
	var location = flag.String("l", "", "Data centre alias (to resolve group and/or network ID)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <server-name>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() != 1 || *hwGroup == "" {
		flag.Usage()
		os.Exit(1)
	}

	client, err := clcv2.NewClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	/* hwGroup may be hex uuid or group name */
	if _, err := hex.DecodeString(*hwGroup); err == nil {
		/* already looks like a HEX ID */
	} else if *location == "" {
		exit.Errorf("Need a location argument (-l) if not using a HW Group UUID (%s)", *hwGroup)
	} else {
		fmt.Printf("Resolving ID of Hardware Group %q ...\n", *hwGroup)

		if group, err := client.GetGroupByName(*hwGroup, *location); err != nil {
			exit.Errorf("Failed to resolve group name %q: %s", *hwGroup, err)
		} else if group == nil {
			exit.Errorf("No group named %q was found on %s", *hwGroup, *location)
		} else {
			*hwGroup = group.Id
		}
	}

	statusId, err := client.RestoreServer(flag.Arg(0), *hwGroup)
	if err != nil {
		exit.Fatalf("Failed to restore server %s: %s", flag.Arg(0), err)
	}

	fmt.Println("Request ID for restoring server:", statusId)
}
