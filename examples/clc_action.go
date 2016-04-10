// Multi-function script for use with CLC servers and hardware groups.
package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path"
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
		{"ip", "print IP addresses of a server, or of all servers in a group"},
		{"on", "power on server (or resume from paused state)"},
		{"off", "power off server"},
		{"shutdown", "OS-level shutdown followed by power-off for server"},
		{"pause", "pause server"},
		{"reset", "perform forced power-cycle on server"},
		{"reboot", "reboot server"},
		{"snapshot", "snapshot server"},
		{"delsnapshot", "delete server snapshot"},
		{"revert", "revert server to snapshot state"},
		{"archive", "archive the server/group"},
		{"delete", "delete server/group (CAUTION)"},
		{"help", "print this help screen"},
	} {
		fmt.Fprintf(os.Stderr, "\t%-10s %s\n", r[0], r[1])
	}
	fmt.Fprintf(os.Stderr, "\n")

	flag.PrintDefaults()
	os.Exit(0)
}

func main() {
	var (
		location       = flag.String("l", "", "Location to use for <Group-Name>")
		intvl          = flag.Duration("i", 20*time.Second, "Poll interval for status updates (use 0 to disable)")
		handlingServer bool   // what to act on
		action, where  string // what to do and where
		reqID          string // request ID of the action
	)

	flag.Usage = usage
	flag.Parse()

	if flag.NArg() >= 2 {
		action, where = flag.Arg(0), flag.Arg(1)
	} else if flag.NArg() == 1 {
		action = flag.Arg(0)
		switch action {
		case "show":
			if *location == "" {
				exit.Errorf("Showing group details requires location (-l) argument.")
			}
		case "help":
			usage()
		case "ip", "on", "off", "shutdown", "pause", "reset", "reboot", "snapshot",
			"delsnapshot", "revert", "archive", "delete":
			/* FIXME: use map for usage, and use keys here, i.e. _, ok := map[action] */
			exit.Errorf("Action %q reaquires an argument (try -h).", action)
		default:
			exit.Errorf("Unsupported action %q", action)
		}
	} else {
		usage()
	}

	client, err := clcv2.NewClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

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
			exit.Errorf("No group named %q was found on %s", where, *location)
		} else {
			where = group.Id
		}
	} else if *location == "" {
		exit.Errorf("%q looks like a group name - need a location (-l argument) to resolve it.", where)
	} else {
		exit.Errorf("Unable to determine whether %q is a server or a group", where)
	}

	if handlingServer { /* Server Action */
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
		}
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

	fmt.Printf("Request ID for %q action: %s\n", action, reqID)
	client.PollStatus(reqID, *intvl)
}

// Print server IP(s)
// @client:    authenticated CLCv2 Client
// @servname:  server name
func printServerIP(client *clcv2.Client, servname string) {
	ips, err := client.GetServerIPs(servname)
	if err != nil {
		exit.Fatalf("Failed get server %q IPs: %s", servname, err)
	}

	fmt.Printf("%-20s %s\n", servname+":", strings.Join(ips, ", "))
}

// Print group hierarchy starting at @g, using initial indentation @indent.
func printGroupIPs(client *clcv2.Client, root *clcv2.Group) {
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
func showServers(client *clcv2.Client, servnames ...string) {

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
func showServer(client *clcv2.Client, servname string) {
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
func showGroup(client *clcv2.Client, root *clcv2.Group) {
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
