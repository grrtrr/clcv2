/*
 * Simplified interface for cloning an existing server.
 */
package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
)

func main() {
	var seed = flag.String("s", "CLONE", "The 4-6 character seed for the name of the cloned server")
	var hwGroup = flag.String("g", "", "UUID or name (if unique) of the HW group to add this server to")
	var location = flag.String("l", "", "Data centre alias (to resolve group and/or network ID)")
	var primDNS = flag.String("dns1", "8.8.8.8", "Primary DNS to use")
	var secDNS = flag.String("dns2", "8.8.4.4", "Secondary DNS to use")
	var numCpu = flag.Int("cpu", 0, "Number of Cpus to use (if different from source VM)")
	var memGB = flag.Int("memory", 0, "Amount of memory in GB (if different from source VM")

	var desc = flag.String("t", "", "Description of the cloned server")
	var net = flag.String("net", "", "ID or name of the Network to use (if different from source)")

	var extraDrv = flag.Int("drive", 0, "Extra storage (in GB) to add to server as a raw disk")

	var ttl = flag.Duration("ttl", 0, "Time span (counting from time of creation) until server gets deleted")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options] <Source-Server-Name>\n", path.Base(os.Args[0]))
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

	// First get the details of the source server
	log.Printf("Obtaining details of %s ...", flag.Arg(0))
	src, err := client.GetServer(flag.Arg(0))
	if err != nil {
		exit.Fatalf("failed to list details of source server %q: %s", flag.Arg(0), err)
	}

	// We need the credentials, too
	log.Printf("Obtaining %s credentials ...", src.Name)
	credentials, err := client.GetServerCredentials(src.Name)
	if err != nil {
		exit.Fatalf("failed to obtain the credentials of server %q: %s", src.Name, err)
	}

	req := clcv2.CreateServerReq{
		Name:                 *seed,
		Cpu:                  src.Details.Cpu,
		MemoryGB:             src.Details.MemoryMb >> 10,
		Description:          fmt.Sprintf("%s (cloned from %s)", src.Description, src.Name),
		GroupId:              src.GroupId,
		SourceServerId:       src.Name,
		PrimaryDns:           *primDNS,
		SecondaryDns:         *secDNS,
		Password:             credentials.Password,
		SourceServerPassword: credentials.Password,

		Type: src.Type,
	}

	if *numCpu != 0 {
		req.Cpu = *numCpu
	}
	if *memGB != 0 {
		req.MemoryGB = *memGB
	}
	if *desc != "" {
		req.Description = *desc
	}
	if *extraDrv != 0 {
		req.AdditionalDisks = append(req.AdditionalDisks,
			clcv2.ServerAdditionalDisk{SizeGB: uint32(*extraDrv), Type: "raw"})
	}
	if *ttl != 0 { /* Date/time that the server should be deleted. */
		req.Ttl = new(time.Time)
		*req.Ttl = time.Now().Add(*ttl)
	}

	/* hwGroup may be hex uuid or group name */
	if *hwGroup != "" {
		req.GroupId = *hwGroup

		if _, err := hex.DecodeString(*hwGroup); err != nil {
			fmt.Printf("Resolving ID of Hardware Group %q ...\n", *hwGroup)

			if group, err := client.GetGroupByName(*hwGroup, *location); err != nil {
				exit.Errorf("failed to resolve group name %q: %s", *hwGroup, err)
			} else if group == nil {
				exit.Errorf("No group named %q was found in %s", *hwGroup, *location)
			} else {
				req.GroupId = group.Id
			}
		}
	}

	/* net is supposed to be a (hex) ID, but allow network names, too */
	if *net == "" {
		log.Printf("Getting IP networks of %s ...", src.Name)

		if nets, err := client.GetServerNets(src); err != nil {
			exit.Fatalf("failed to query networks of %s: %s", src.Name, err)
		} else if len(nets) == 0 {
			exit.Fatalf("no network information found for %s", src.Name)
		} else if len(nets) != 1 {
			exit.Errorf("please specify which network to use (%s uses %d)", src.Name, len(nets))
		} else {
			req.NetworkId = nets[0].Id
		}
	} else if _, err := hex.DecodeString(*net); err == nil {
		/* already looks like a HEX ID */
	} else if *location == "" {
		exit.Errorf("Need a location argument (-l) if not using a network ID (%s)", *net)
	} else {
		fmt.Printf("Resolving network id of %q ...\n", *net)

		if netw, err := client.GetNetworkIdByName(*net, *location); err != nil {
			exit.Errorf("failed to resolve network name %q: %s", *net, err)
		} else if netw == nil {
			exit.Errorf("No network named %q was found in %s", *net, *location)
		} else {
			req.NetworkId = netw.Id
		}
	}

	name, status, err := client.CreateServer(&req)
	if err != nil {
		exit.Fatalf("failed to create server: %s", err)
	}

	fmt.Printf("New server name: %s\n", name)
	fmt.Printf("Server Password: \"%s\"\n", credentials.Password)
	fmt.Printf("Status Id:       %s\n", status)
}
