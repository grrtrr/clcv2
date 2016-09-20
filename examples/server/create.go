/*
 * Create a new server. Does not support creation of bare-metal servers (yet).
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

	humanize "github.com/dustin/go-humanize"
	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
	"github.com/olekukonko/tablewriter"
)

func main() {
	var hwGroup = flag.String("g", "", "UUID or name (if unique) of the HW group to add this server to")
	var location = flag.String("l", "", "Data centre alias (to resolve group and/or network ID)")
	var srcServer = flag.String("src", "", "The name of a source-server, or a template, to create from")
	var srcPass = flag.String("srcPass", "", "When cloning from a source-server, use this password")
	var seed = flag.String("s", "AUTO", "The seed for the server name")
	var desc = flag.String("t", "", "Description of the server")

	var net = flag.String("net", "", "ID or name of the Network to use")
	var primDNS = flag.String("dns1", "8.8.8.8", "Primary DNS to use")
	var secDNS = flag.String("dns2", "8.8.4.4", "Secondary DNS to use")
	var password = flag.String("pass", "", "Desired password. Leave blank to auto-generate")

	var extraDrv = flag.Int("drive", 0, "Extra storage (in GB) to add to server as a raw disk")
	var numCpu = flag.Int("cpu", 1, "Number of Cpus to use")
	var memGB = flag.Int("memory", 4, "Amount of memory in GB")
	var serverType = flag.String("type", "standard", "The type of server to create (standard, hyperscale, or bareMetal)")
	var ttl = flag.Duration("ttl", 0, "Time span (counting from time of creation) until server gets deleted")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	if *hwGroup == "" {
		flag.Usage()
		os.Exit(0)
	}

	client, err := clcv2.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	/* hwGroup may be hex uuid or group name */
	if _, err := hex.DecodeString(*hwGroup); err != nil {
		fmt.Printf("Resolving ID of Hardware Group %q ...\n", *hwGroup)

		if group, err := client.GetGroupByName(*hwGroup, *location); err != nil {
			exit.Errorf("failed to resolve group name %q: %s", *hwGroup, err)
		} else if group == nil {
			exit.Errorf("No group named %q was found in %s", *hwGroup, *location)
		} else {
			*hwGroup = group.Id
		}
	}

	/* net is supposed to be a (hex) ID, but allow network names, too */
	if *net != "" {
		if _, err := hex.DecodeString(*net); err == nil {
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
			clcv2.ServerAdditionalDisk{SizeGB: uint32(*extraDrv), Type: "raw"})
	}

	/* Date/time that the server should be deleted. */
	if *ttl != 0 {
		req.Ttl = new(time.Time)
		*req.Ttl = time.Now().Add(*ttl)
	}

	name, reqID, err := client.CreateServer(&req)
	if err != nil {
		exit.Fatalf("failed to create server: %s", err)
	}

	log.Printf("New server name: %s", name)
	log.Printf("Status Id: %s", reqID)

	client.PollStatus(reqID, 5*time.Second)

	// Print details after job completes
	showServer(client, name)
}

// Show details of a single server (taken from clc_action.go)
// @client:    authenticated CLCv2 Client
// @servname:  server name
func showServer(client *clcv2.CLIClient, servname string) {
	server, err := client.GetServer(servname)
	if err != nil {
		exit.Fatalf("failed to list details of server %q: %s", servname, err)
	}

	grp, err := client.GetGroup(server.GroupId)
	if err != nil {
		exit.Fatalf("failed to resolve group UUID: %s", err)
	}

	/* First public, then private */
	IPs := []string{}
	for _, ip := range server.Details.IpAddresses {
		if ip.Public != "" {
			IPs = append(IPs, ip.Public)
		}
	}
	for _, ip := range server.Details.IpAddresses {
		if ip.Internal != "" {
			IPs = append(IPs, ip.Internal)
		}
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoWrapText(true)

	// CPU, Memory, IP and Power status are not filled in until the server reaches 'active' state.
	if server.Status == "active" {
		table.SetHeader([]string{
			"Name", "Group", "Description", "OS",
			"CPU", "Mem", "IP", "Power",
			"Last Change",
		})

	} else {
		table.SetHeader([]string{
			"Name", "Group", "Description", "OS",
			"Status",
			"Owner", "Last Change",
		})
	}

	modifiedStr := humanize.Time(server.ChangeInfo.ModifiedDate)
	/* The ModifiedBy field can be an email address, or an API Key (hex string) */
	if _, err := hex.DecodeString(server.ChangeInfo.ModifiedBy); err == nil {
		modifiedStr += " via API Key"
	} else if len(server.ChangeInfo.ModifiedBy) > 6 {
		modifiedStr += " by " + server.ChangeInfo.ModifiedBy[:6]
	} else {
		modifiedStr += " by " + server.ChangeInfo.ModifiedBy
	}

	if server.Status == "active" {
		table.Append([]string{
			server.Name, grp.Name, server.Description, server.OsType,
			fmt.Sprint(server.Details.Cpu), fmt.Sprintf("%d G", server.Details.MemoryMb/1024), strings.Join(IPs, " "), server.Details.PowerState,
			modifiedStr,
		})
	} else {
		table.Append([]string{
			server.Name, grp.Name, server.Description, server.OsType,
			server.Status,
			server.ChangeInfo.CreatedBy, modifiedStr,
		})
	}
	table.Render()

	// Disks
	if len(server.Details.Disks) > 0 {
		fmt.Printf("\nDisks of %s (total storage: %d GB)\n", server.Name, server.Details.StorageGb)
		table = tablewriter.NewWriter(os.Stdout)
		table.SetAutoFormatHeaders(false)
		table.SetAlignment(tablewriter.ALIGN_RIGHT)
		table.SetAutoWrapText(true)

		table.SetHeader([]string{"Disk ID", "Disk Size/GB", "Paths"})
		for _, d := range server.Details.Disks {
			table.Append([]string{d.Id, fmt.Sprint(d.SizeGB), strings.Join(d.PartitionPaths, ", ")})
		}
		table.Render()
	}

	// Partitions
	if len(server.Details.Partitions) > 0 {
		fmt.Printf("\nPartitions of %s:\n", server.Name)
		table = tablewriter.NewWriter(os.Stdout)
		table.SetAutoFormatHeaders(false)
		table.SetAlignment(tablewriter.ALIGN_RIGHT)
		table.SetAutoWrapText(true)

		table.SetHeader([]string{"Partition Path", "Partition Size/GB"})
		for _, p := range server.Details.Partitions {
			table.Append([]string{p.Path, fmt.Sprintf("%.1f", p.SizeGB)})
		}
		table.Render()
	}

	// Snapshots
	if len(server.Details.Snapshots) > 0 {
		fmt.Println()

		table = tablewriter.NewWriter(os.Stdout)
		table.SetAutoFormatHeaders(false)
		table.SetAlignment(tablewriter.ALIGN_CENTRE)
		table.SetAutoWrapText(true)

		table.SetHeader([]string{fmt.Sprintf("Snapshots of %s", server.Name)})
		for _, s := range server.Details.Snapshots {
			table.Append([]string{s.Name})
		}
		table.Render()
	}
}
