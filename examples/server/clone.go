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
	"strings"
	"time"

	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
	"github.com/olekukonko/tablewriter"
)

func main() {
	var net = flag.String("net", "", "ID or name of the Network to use (if different from source)")
	var hwGroup = flag.String("g", "", "UUID or name (if unique) of the HW group to clone this server into")
	var primDNS = flag.String("dns1", "8.8.8.8", "Primary DNS to use")
	var secDNS = flag.String("dns2", "8.8.4.4", "Secondary DNS to use")
	var numCpu = flag.Int("cpu", 0, "Number of Cpus to use (if different from source VM)")
	var memGB = flag.Int("memory", 0, "Amount of memory in GB (if different from source VM")
	var seed = flag.String("name", "", "The 4-6 character seed for the name of the cloned server")
	var desc = flag.String("desc", "", "Description of the cloned server")
	var ttl = flag.Duration("ttl", 0, "Time span (counting from time of creation) until server gets deleted")
	var extraDrv = flag.Int("drive", 0, "Extra storage (in GB) to add to server as a raw disk")
	var wasStopped bool
	var maxAttempts = 1
	var url, reqID string

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
	log.Printf("Obtaining details of source server %s ...", flag.Arg(0))
	src, err := client.GetServer(flag.Arg(0))
	if err != nil {
		exit.Fatalf("failed to list details of source server %q: %s", flag.Arg(0), err)
	}

	if wasStopped = src.Details.PowerState == "stopped"; wasStopped {
		// The source server must be powered on
		log.Printf("%s is powered-off - powering on ...", src.Name)
		reqID, err := client.PowerOnServer(src.Name)
		if err != nil {
			exit.Fatalf("failed to power on source server %s: %s", src.Name, err)
		}
		log.Printf("Waiting for %s to power on (status ID: %s) ...", src.Name, reqID)
		if _, err = client.AwaitCompletion(reqID); err != nil {
			exit.Fatalf("failed to await completion of %s: %s", reqID, err)
		}
		// When the server is being powered on, it can take up to 5 minutes until
		// the backend is able to clone it; it requires the server to be fully booted.
		maxAttempts = 5
		time.Sleep(1 * time.Minute)
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
		GroupId:              src.GroupId,
		SourceServerId:       src.Name,
		PrimaryDns:           *primDNS,
		SecondaryDns:         *secDNS,
		Password:             credentials.Password,
		SourceServerPassword: credentials.Password,

		Type: src.Type,
	}

	if *seed == "" {
		if l := len(src.Name); l > 10 { // use same naming as original by default
			req.Name = src.Name[7 : l-1]
		} else {
			req.Name = "CLONE"
		}
	}

	if *numCpu != 0 {
		req.Cpu = *numCpu
	}
	if *memGB != 0 {
		req.MemoryGB = *memGB
	}

	if *desc != "" {
		req.Description = *desc
	} else if src.Description == "" {
		req.Description = fmt.Sprintf("Clone of %s", src.Name)
	} else {
		req.Description = fmt.Sprintf("%s (cloned from %s)", src.Description, src.Name)
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
			log.Printf("Resolving ID of Hardware Group %q in %s ...", *hwGroup, src.LocationId)

			if group, err := client.GetGroupByName(*hwGroup, src.LocationId); err != nil {
				exit.Fatalf("failed to resolve group name %q: %s", *hwGroup, err)
			} else if group == nil {
				exit.Errorf("No group named %q was found in %s", *hwGroup, src.LocationId)
			} else {
				req.GroupId = group.Id
			}
		}
	}

	/* net is supposed to be a (hex) ID, but allow network names, too */
	if *net == "" {
		log.Printf("Determining network ID used by %s ...", src.Name)

		nets, err := client.GetServerNets(src)
		if err != nil {
			exit.Fatalf("failed to query networks of %s: %s", src.Name, err)
		}

		if len(nets) == 0 {
			// No network information found for the server, even though it has an IP.
			// This can happen when the server is owned by a sub-account, and uses a
			// network that is owned by the parent account. In this case, the sub-account
			// is prevented from querying details of the parent account, due to insufficient
			// permission.
			if parentAccount := client.RegisteredAccountAlias(); client.AccountAlias != parentAccount {
				var savedAlias = client.AccountAlias

				log.Printf("Network ID not visible under %q account - trying %q instead ...", savedAlias, parentAccount)
				client.AccountAlias = parentAccount
				if nets, err = client.GetServerNets(src); err != nil {
					exit.Fatalf("failed to query networks of %s using %q account: %s", src.Name, parentAccount, err)
				}
				// Restore Account Alias for remainder of program
				client.AccountAlias = savedAlias
			}
			if len(nets) == 0 {
				log.Printf("Unable to determine Network ID - querying %s deployable networks ...", src.LocationId)
				capa, err := client.GetDeploymentCapabilities(src.LocationId)
				if err != nil {
					exit.Fatalf("failed to determine %s Deployment Capabilities: %s", src.LocationId, err)
				}
				fmt.Println("Please specify the network ID for the clone manually via -net, using this information:")
				table := tablewriter.NewWriter(os.Stdout)
				table.SetAutoFormatHeaders(false)
				table.SetAlignment(tablewriter.ALIGN_LEFT)
				table.SetAutoWrapText(false)

				table.SetHeader([]string{"Name", "Type", "Account", "Network ID"})
				for _, net := range capa.DeployableNetworks {
					table.Append([]string{net.Name, net.Type, net.AccountID, net.NetworkId})
				}

				table.Render()
				os.Exit(0)
			}
		}

		if len(nets) != 1 {
			// FIXME: print server networks
			exit.Errorf("please specify which network to use (%s uses %d)", src.Name, len(nets))
		} else {
			log.Printf("Using %s network %s, with gateway %s", nets[0].Type, nets[0].Cidr, nets[0].Gateway)
			req.NetworkId = nets[0].Id
		}
	} else if _, err := hex.DecodeString(*net); err != nil { // not a HEX ID, treat as group name
		log.Printf("Resolving network ID of %q in %s ...", *net, src.LocationId)

		if netw, err := client.GetNetworkIdByName(*net, src.LocationId); err != nil {
			exit.Fatalf("failed to resolve network name %q: %s", *net, err)
		} else if netw == nil {
			exit.Fatalf("unable to resolve network name %q in %s - maybe use hex ID?", *net, src.LocationId)
		} else {
			req.NetworkId = netw.Id
		}
	} else { // HEX ID, use directly
		req.NetworkId = *net
	}

	log.Printf("Cloning %s ...", src.Name)
	for i := 1; ; i++ {
		url, reqID, err = client.CreateServer(&req)
		if err == nil || i == maxAttempts || strings.Index(err.Error(), "body.sourceServerId") > 0 {
			break
		}
		log.Printf("attempt %d/%d failed (%s) - retrying ...", i, maxAttempts, strings.TrimSpace(err.Error()))
		time.Sleep(1 * time.Minute)
	}
	if err != nil {
		exit.Fatalf("failed to create server: %s", err)
	}

	status, err := client.PollStatus(reqID, 5*time.Second)
	if err != nil {
		exit.Fatalf("failed to poll %s status: %s", reqID, err)
	}

	server, err := client.GetServerByURI(url)
	if err != nil {
		log.Fatalf("failed to query server details at %s: %s", url, err)
	} else if status == clcv2.Failed {
		exit.Fatalf("failed to clone %s (will show up as 'under construction')", server.Name)
	}
	log.Printf("New server name: %s\n", server.Name)
	log.Printf("Server Password: \"%s\"\n", credentials.Password)
}
