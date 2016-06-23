/*
 * Remove a secondary network adapter from a given server.
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
	var net      = flag.String("net", "", "ID or name of the Network to use (REQUIRED)")
	var location = flag.String("l",   "", "Data centre alias (to resolve network name if not using hex ID)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options] <Server-Name>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() != 1 || *net == "" {
		flag.Usage()
		os.Exit(0)
	}

	client, err := clcv2.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	/* net is supposed to be a (hex) ID, but allow network names, too */
	if *net != "" {
		if _, err := hex.DecodeString(*net); err == nil {
			/* already looks like a HEX ID */
		} else if  *location == "" {
			exit.Errorf("Need a location argument (-l) if not using a network ID (%s)", *net)
		} else {
			fmt.Printf("Resolving network id of %q ...\n", *net)

			if netw, err := client.GetNetworkIdByName(*net, *location); err != nil {
				exit.Errorf("Failed to resolve network name %q: %s", *net, err)
			} else if netw == nil {
				exit.Errorf("No network named %q was found in %s", *net, *location)
			} else {
				*net = netw.Id
			}
		}
	}

	if err = client.ServerDelNic(flag.Arg(0), *net); err != nil {
		exit.Fatalf("Failed to remove NIC from %s: %s", flag.Arg(0), err)
	}

	fmt.Printf("Successfully removed secondary NIC from %s.\n", flag.Arg(0))

}
