/*
 * Create a new standard or hyperscale server.
 * Does not support creation of bare-metal servers (yet).
 */
package main

import (
	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
	"encoding/hex"
	"time"
	"path"
	"flag"
	"fmt"
	"os"
)

func main() {
	var hwGroup    = flag.String("g",       "",      "UUID or name (if unique) of the HW group to add this server to")
	var location   = flag.String("l",       "",      "Data centre alias (only used in conjunction with -g)")
	var srcServer  = flag.String("src",     "",      "The name of a source-server, or a template, to create from")
	var srcPass    = flag.String("srcPass", "",      "When cloning from a source-server, use this password")
	var seed       = flag.String("s",       "AUTO",  "The seed for the server name")
	var desc       = flag.String("t",       "",      "Description of the server")

	var net        = flag.String("net",  "",         "ID or name of the Network to use")
	var primDNS    = flag.String("dns1", "8.8.8.8",  "Primary DNS to use")
	var secDNS     = flag.String("dns2", "8.8.4.4",  "Secondary DNS to use")
	var password   = flag.String("pass", "",         "Desired password. Leave blank to auto-generate")

	var extraDrv   = flag.Int("drive",  0,           "Extra storage (in GB) to add to server as a raw disk")
	var numCpu     = flag.Int("cpu",    1,           "Number of Cpus to use")
	var memGB      = flag.Int("memory", 4,           "Amount of memory in GB")
	var serverType = flag.String("type", "standard", "The type of server to create (standard, hyperscale, or bareMetal)")
	var storType   = flag.String("level", "premium", "Data storage service level (standard or premium)")
	var ttl        = flag.Duration("ttl",  0,        "Time span (counting from time of creation) until server gets deleted")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	if *hwGroup == "" {
		flag.Usage()
		os.Exit(0)
	}

	client, err := clcv2.NewClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	/* hwGroup may be hex uuid or group name */
	if _, err := hex.DecodeString(*hwGroup); err != nil {
		fmt.Printf("Resolving ID of Hardware Group %q ...\n", *hwGroup)

		if group, err := client.GetGroupByName(*hwGroup, *location); err != nil {
			exit.Errorf("Failed to resolve group name %q: %s", *hwGroup, err)
		} else if group == nil {
			exit.Errorf("No group named %q was found on %s", *hwGroup, *location)
		} else {
			*hwGroup = group.Id
		}
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
				exit.Errorf("No network named %q was found on %s", *net, *location)
			} else {
				*net = netw.Id
			}
		}
	}

	req := clcv2.CreateServerReq{
		// Name of the server to create. Alphanumeric characters and dashes only.
		Name: *seed,

		// User-defined description of this server
		Description: *desc,

		// ID of the parent HW group.
		GroupId: *hwGroup,

		// ID of the server to use a source. May be the ID of a srcServer, or when cloning, an existing server ID.
		SourceServerId: *srcServer,

		// The primary DNS to set on the server
		PrimaryDns: *primDNS,

		// The secondary DNS to set on the server
		SecondaryDns: *secDNS,

		// ID of the network to which to deploy the server.
		NetworkId: *net,

		// Password of administrator or root user on server.
		Password: *password,

		// Password of the source server, used only when creating a clone from an existing server.
		SourceServerPassword: *srcPass,

		// Number of processors to configure the server with (1-16)
		Cpu: *numCpu,

		// Number of GB of memory to configure the server with (1-128)
		MemoryGB: *memGB,

		// Whether to create a 'standard', 'hyperscale', or 'bareMetal' server
		Type: *serverType,

		// For standard servers, whether to use standard or premium storage.
		StorageType: *storType,

		// FIXME: the following are not populated in this request:
		// - IpAddress
		// - IsManagedOs
		// - IsManagedBackup
		// - AntiAffinityPolicyId
		// - CpuAutoscalePolicyId
		// - CustomFields
		// - Packages
		//
		// The following items relevant specific to bare-metal servers are also ignored:
		// - ConfigurationId
		// - OsType
	}

	if *extraDrv != 0 {
		req.AdditionalDisks = append(req.AdditionalDisks,
					     clcv2.ServerDisk{SizeGB: *extraDrv, Type: "raw"})
	}

	/* Date/time that the server should be deleted. */
	if *ttl != 0 {
		req.Ttl = new(time.Time)
		*req.Ttl = time.Now().Add(*ttl)
	}

	name, status, err := client.CreateServer(&req)
	if err != nil {
		exit.Fatalf("Failed to create server: %s", err)
	}

	fmt.Printf("New server name: %s\n", name)
	fmt.Printf("Status Id:       %s\n", status)
}
