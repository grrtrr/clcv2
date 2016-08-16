/*
 * Add a secondary network adapter to a given server.
 * Use this API operation when you need to add a secondary network adapter to a server.
 * Users have the option to specify an IP address to assign to the server; otherwise the
 * first available IP address in the network will be assigned. Up to four total network
 * adapters can be attached to a server (i.e. a total of 3 secondary adapters).
 * In addition, only one IP address per secondary network can be associated with a server.
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
	var net      = flag.String("net", "", "ID or name of the Network to use")
	var location = flag.String("l",   "", "Data centre alias (to resolve network name if not using hex ID)")
	var ip       = flag.String("ip",  "", "IP address on -net (optional, default is automatic assignment)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options] <Server-Name>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() != 1 {
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
				exit.Errorf("failed to resolve network name %q: %s", *net, err)
			} else if netw == nil {
				exit.Errorf("No network named %q was found in %s", *net, *location)
			} else {
				*net = netw.Id
			}
		}
	}

	if err = client.ServerAddNic(flag.Arg(0), *net, *ip); err != nil {
		exit.Fatalf("failed to add NIC to %s: %s", flag.Arg(0), err)
	}

	fmt.Printf("Successfully added a secondary NIC to %s.\n", flag.Arg(0))
}
