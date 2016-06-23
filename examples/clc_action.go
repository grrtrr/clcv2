// Multi-function script for use with CLC servers and hardware groups.
package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/clcv2/utils"
	"github.com/grrtrr/exit"
	"github.com/olekukonko/tablewriter"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage:\n")
	fmt.Fprintf(os.Stderr, "\t%s [options]      <action>  <Server-Name|Group-UUID>\n", path.Base(os.Args[0]))
	fmt.Fprintf(os.Stderr, "\t%s -l <Location>  <action>  <Group-Name>\n\n", path.Base(os.Args[0]))

	for _, r := range [][]string{
		{"show", "show current status of server/group (group requires -l to be set)"},
		{"templates", "show the templates for a given region (requires -l to be set)"},
		{"networks", "show the networks available in a given region (requires -l to be set)"},
		{"ip", "print IP addresses of a server, or of all servers in a group"},
		{"on", "power on server (or resume from paused state)"},
		{"off", "power off server"},
		{"shutdown", "OS-level shutdown followed by power-off for server"},
		{"pause", "pause server"},
		{"reset", "perform forced power-cycle on server"},
		{"reboot", "reboot server"},
		{"rawdisk", "<sizeGB> - add storage to server"},
		{"snapshot", "snapshot server"},
		{"delsnapshot", "delete server snapshot"},
		{"revert", "revert server to snapshot state"},
		{"archive", "archive the server/group"},
		{"delete", "delete server/group (CAUTION)"},
		{"help", "print this help screen"},
	} {
		fmt.Fprintf(os.Stderr, "\t%-15s %s\n", r[0], r[1])
	}
	fmt.Fprintf(os.Stderr, "\n")

	flag.PrintDefaults()
	os.Exit(0)
}

func main() {
	var (
		location       = flag.String("l", "", "Location to use for <Group-Name>")
		intvl          = flag.Duration("i", 1*time.Second, "Poll interval for status updates (use 0 to disable)")
		handlingServer bool   // what to act on
		action, where  string // what to do and where
		reqID          string // request ID of the action
	)

	/*
	 * Argument Validation
	 */
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() >= 1 {
		action = flag.Arg(0)
		if flag.NArg() >= 2 {
			where = flag.Arg(1)
		}
	} else {
		usage()
	}

	switch action {
	case "help":
		usage()
	case "networks", "show", "templates":
		if flag.NArg() == 1 && *location == "" {
			exit.Errorf("Action %q requires location (-l argument).", action)
		}
	case "rawdisk":
		handlingServer = true
		if flag.NArg() != 3 {
			exit.Errorf("usage: rawdisk <serverName> <diskGB>")
		}
	case "ip", "on", "off", "shutdown", "pause", "reset", "reboot", "snapshot",
		"delsnapshot", "revert", "archive", "delete":
		/* FIXME: use map for usage, and use keys here, i.e. _, ok := map[action] */
		if where == "" {
			exit.Errorf("Action %q requires an argument (try -h).", action)
		}
	default:
		exit.Errorf("Unsupported action %q", action)
	}

	client, err := clcv2.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	if !handlingServer {
		/*
		 * Decide if arguments refer to a server or a hardware group.
		 */
		if _, err := hex.DecodeString(where); err == nil {
			/* If the first argument decodes as a hex value, assume it is a Hardware Group UUID */
		} else if utils.LooksLikeServerName(where) { /* Starts with a location identifier and is not hex ... */
			handlingServer = true
		} else if *location != "" && where != "" {
			if group, err := client.GetGroupByName(where, *location); err != nil {
				exit.Errorf("Failed to resolve group name %q: %s", where, err)
			} else if group == nil {
				exit.Errorf("No group named %q was found in %s", where, *location)
			} else {
				where = group.Id
			}
		} else if *location == "" {
			exit.Errorf("%q looks like a group name - need a location (-l argument) to resolve it.", where)
		} else {
			exit.Errorf("Unable to determine whether %q is a server or a group", where)
		}
	}

	if action == "templates" { /* where="" - neither server nor group action; print regional templates */
		showTemplates(client, *location)
		os.Exit(0)
	} else if action == "networks" { /* similar, neither server nor group action; print regional networks */
		showNetworks(client, *location)
		os.Exit(0)
	} else if handlingServer { /* Server Action */
		switch action {
		case "ip":
			printServerIP(client, where)
			os.Exit(0)
		case "show":
			// FIXME: deal with multiple servers
			if flag.NArg() == 2 {
				showServer(client, where)
			} else {
				showServers(client, flag.Args()[1:]...)
			}
			os.Exit(0)
		case "rawdisk":
			diskGB, err := strconv.ParseUint(flag.Arg(2), 10, 32)
			if err != nil {
				exit.Errorf("rawdisk: invalid disk size in GB %q for %s", flag.Arg(2), where)
			}
			reqID = addRawDisk(client, where, uint32(diskGB))
		default:
			var serverAction = map[string]func(string) (string, error){
				"on":          client.PowerOnServer,
				"off":         client.PowerOffServer,
				"pause":       client.PauseServer,
				"reset":       client.ResetServer,
				"reboot":      client.RebootServer,
				"shutdown":    client.ShutdownServer,
				"archive":     client.ArchiveServer,
				"delete":      client.DeleteServer,
				"snapshot":    client.SnapshotServer,
				"delsnapshot": client.DeleteSnapshot,
				"revert":      client.RevertToSnapshot,
			}

			/* Long-running commands that return a RequestID */
			handler, ok := serverAction[action]
			if !ok {
				exit.Fatalf("Unsupported server action %s", action)
			}

			reqID, err = handler(where)
			if err != nil {
				exit.Fatalf("Server command %q failed: %s", action, err)
			}
		}
	} else if action == "show" || action == "ip" {
		/* Printing group trees: requires to resolve the root first. */
		var start *clcv2.Group

		if *location == "" {
			exit.Errorf("Location argument (-l) is required in order to traverse nested groups.")
		}

		root, err := client.GetGroups(*location)
		if err != nil {
			exit.Fatalf("Failed to look up groups at %s: %s", *location, err)
		}

		start = &root
		if where != "" {
			start = clcv2.FindGroupNode(start, func(g *clcv2.Group) bool {
				return g.Id == where
			})
			if start == nil {
				exit.Fatalf("Failed to look up UUID %s at %s", where, location)
			}
		}

		switch action {
		case "show":
			showGroup(client, start)
		case "ip":
			printGroupIPs(client, start)
		}
		os.Exit(0)
	} else {
		/* Other Group Action */
		switch action {
		case "archive":
			reqID, err = client.ArchiveGroup(where)
		case "delete":
			reqID, err = client.DeleteGroup(where)
		default:
			exit.Errorf("Unsupported group action %q", action)
		}
		if err != nil {
			exit.Fatalf("Group command %q failed: %s", action, err)
		}
	}

	client.PollStatus(reqID, *intvl)
}

// showTemplates prints the templates available in @region
func showTemplates(client *clcv2.CLIClient, region string) {
	capa, err := client.GetDeploymentCapabilities(region)
	if err != nil {
		exit.Fatalf("Failed to query deployment capabilities of %s: %s", region, err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoWrapText(false)

	/* Note: not displaying ReservedDrivePaths and DrivePathLength here, I don't understand their use. */
	/* Note: not listing Capabilities here, since the table gets too large for a single screen */
	table.SetHeader([]string{"Template Name", "Description", "OS", "Storage"})

	for _, tpl := range capa.Templates {
		table.Append([]string{tpl.Name, tpl.Description, tpl.OsType, fmt.Sprintf("%d GB", tpl.StorageSizeGB)})
	}
	table.Render()
}

// Print server IP(s)
// @client:    authenticated CLCv2 Client
// @servname:  server name
func printServerIP(client *clcv2.CLIClient, servname string) {
	ips, err := client.GetServerIPs(servname)
	if err != nil {
		exit.Fatalf("Failed get server %q IPs: %s", servname, err)
	}

	fmt.Printf("%-20s %s\n", servname+":", strings.Join(ips, ", "))
}

// Print group hierarchy starting at @g, using initial indentation @indent.
func printGroupIPs(client *clcv2.CLIClient, root *clcv2.Group) {
	var serverPrinter = func(g *clcv2.Group, arg interface{}) interface{} {
		var indent = arg.(string)

		if g.Type == "default" {
			fmt.Printf("%s%s/\n", indent, g.Name)
		} else {
			fmt.Printf("%s[%s]/\n", indent, g.Name)
		}

		for _, l := range g.Links {
			if l.Rel == "server" {
				ips, err := client.GetServerIPs(l.Id)
				if err != nil {
					exit.Fatalf("Failed to get IPs of %q in %s: %s", l.Id, g.Name, err)
				}

				servLine := fmt.Sprintf("%s%s", indent+"    ", l.Id)
				fmt.Printf("%-50s %s\n", servLine, strings.Join(ips, ", "))
			}
		}
		return indent + "    "
	}
	clcv2.VisitGroupHierarchy(root, serverPrinter, "")
}

// Condensed details of multiple servers
// @client:    authenticated CLCv2 Client
// @servnames: server names
func showServers(client *clcv2.CLIClient, servnames ...string) {

	truncate := func(s string, maxlen int) string {
		if len(s) >= maxlen {
			s = s[:maxlen]
		}
		return s
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoWrapText(true)

	table.SetHeader([]string{
		"Name", "Description", "OS",
		"IP", "CPU", "Mem", "Storage",
		"Status", "Power", "Last Change",
	})

	for _, servname := range servnames {
		server, err := client.GetServer(servname)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to list details of server %q: %s", servname, err)
			continue
		}

		IPs := []string{}
		for _, ip := range server.Details.IpAddresses {
			if ip.Public != "" {
				IPs = append(IPs, ip.Public)
			}
			if ip.Internal != "" {
				IPs = append(IPs, ip.Internal)
			}
		}

		status := server.Status
		if server.Details.InMaintenanceMode {
			status = "MAINTENANCE"
		}

		desc := server.Description
		if server.IsTemplate {
			desc = "TPL: " + desc
		}

		modifiedStr := humanize.Time(server.ChangeInfo.ModifiedDate)
		/* The ModifiedBy field can be an email address, or an API Key (hex string) */
		if _, err := hex.DecodeString(server.ChangeInfo.ModifiedBy); err == nil {
			modifiedStr += " via API Key"
		} else {
			modifiedStr += " by " + truncate(server.ChangeInfo.ModifiedBy, 6)
		}

		table.Append([]string{
			server.Name, truncate(desc, 30), truncate(server.OsType, 15),
			strings.Join(IPs, " "),
			fmt.Sprint(server.Details.Cpu), fmt.Sprintf("%d G", server.Details.MemoryMb/1024),
			fmt.Sprintf("%d G", server.Details.StorageGb),
			status, server.Details.PowerState, modifiedStr,
		})
	}
	table.Render()
}

// Show details of a single server
// @client:    authenticated CLCv2 Client
// @servname:  server name
func showServer(client *clcv2.CLIClient, servname string) {
	server, err := client.GetServer(servname)
	if err != nil {
		exit.Fatalf("Failed to list details of server %q: %s", servname, err)
	}

	grp, err := client.GetGroup(server.GroupId)
	if err != nil {
		exit.Fatalf("Failed to resolve group UUID: %s", err)
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

	table.SetHeader([]string{
		"Name", "Group", "Description", "OS",
		"CPU", "Mem",
		"IP", "Power", "Last Change",
	})

	modifiedStr := humanize.Time(server.ChangeInfo.ModifiedDate)
	/* The ModifiedBy field can be an email address, or an API Key (hex string) */
	if _, err := hex.DecodeString(server.ChangeInfo.ModifiedBy); err == nil {
		modifiedStr += " via API Key"
	} else if len(server.ChangeInfo.ModifiedBy) > 6 {
		modifiedStr += " by " + server.ChangeInfo.ModifiedBy[:6]
	} else {
		modifiedStr += " by " + server.ChangeInfo.ModifiedBy
	}

	table.Append([]string{
		server.Name, grp.Name, server.Description, server.OsType,
		fmt.Sprint(server.Details.Cpu), fmt.Sprintf("%d G", server.Details.MemoryMb/1024),
		strings.Join(IPs, " "),
		server.Details.PowerState, modifiedStr,
	})
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

// Show nested group details
// @client: authenticated CLCv2 Client
// @root:   root group node to start at
func showGroup(client *clcv2.CLIClient, root *clcv2.Group) {
	var groupPrinter = func(g *clcv2.Group, arg interface{}) interface{} {
		var indent = arg.(string)
		var groupLine string

		if g.Type == "default" {
			groupLine = fmt.Sprintf("%s%s/", indent, g.Name)
		} else {
			groupLine = fmt.Sprintf("%s[%s]/", indent, g.Name)
		}
		fmt.Printf("%-70s %s\n", groupLine, g.Id)

		/* Nested entries: */
		for _, l := range g.Links {
			if l.Rel == "server" {
				fmt.Printf("%s", indent+"    ")
				fmt.Printf("%s\n", l.Id)
			}
		}
		return indent + "    "
	}
	clcv2.VisitGroupHierarchy(root, groupPrinter, "")
}

// addRawDisk adds storage in form of a raw disk to a server
// @client:   authenticated CLCv2 client
// @servname: server name
// @diskGB:   amount of storage in GB to add to @servname
func addRawDisk(client *clcv2.CLIClient, servname string, diskGB uint32) (statusId string) {
	/* First get the list of disks */
	server, err := client.GetServer(servname)
	if err != nil {
		exit.Fatalf("Failed to list details of server %q: %s", servname, err)
	}

	disks := make([]clcv2.ServerAdditionalDisk, len(server.Details.Disks))
	for i := range server.Details.Disks {
		disks[i] = clcv2.ServerAdditionalDisk{
			Id:     server.Details.Disks[i].Id,
			SizeGB: server.Details.Disks[i].SizeGB,
		}
	}

	statusId, err = client.ServerSetDisks(servname, append(disks,
		clcv2.ServerAdditionalDisk{
			SizeGB: diskGB,
			Type:   "raw",
		}))
	if err != nil {
		exit.Fatalf("Failed to update the disk configuration on %q: %s", servname, err)
	}
	return statusId
}

// showNetworks shows available networks in data centre location @location.
func showNetworks(client *clcv2.CLIClient, location string) {
	networks, err := client.GetNetworks(location)
	if err != nil {
		exit.Fatalf("Failed to list networks in %s: %s", location, err)
	}

	if len(networks) == 0 {
		println("Empty result.")
	} else {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoFormatHeaders(false)
		table.SetAlignment(tablewriter.ALIGN_RIGHT)
		table.SetAutoWrapText(false)

		table.SetHeader([]string{"CIDR", "Gateway", "VLAN", "Name", "Description", "Type", "ID"})
		for _, l := range networks {
			table.Append([]string{l.Cidr, l.Gateway, fmt.Sprint(l.Vlan), l.Name, l.Description, l.Type, l.Id})
		}
		table.Render()
	}
}
