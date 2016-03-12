/*
 * Rewrite of the 'clc_action' bash script into go.
 */
package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

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
		{"on", "power on server (or resume from paused state)"},
		{"off", "power off server"},
		{"shutdown", "OS-level shutdown followed by power-off for server"},
		{"pause", "pause server"},
		{"reset", "perform forced power-cycle on server"},
		{"reboot", "reboot server"},
		{"snapshot", "snapshot server"},
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
		location      = flag.String("l", "", "Location to use for <Group-Name>")
		server_action bool   // what to act on
		action, where string // what to do and where
		reqId         string // request ID of the action
	)

	flag.Usage = usage
	flag.Parse()

	if flag.NArg() == 2 {
		action, where = flag.Arg(0), flag.Arg(1)
	} else if flag.NArg() == 1 {
		// FIXME: make switch statement, and implement
		// - show
		// - help
		// - status
		if flag.Arg(0) == "show" && *location == "" {
			exit.Errorf("Showing group details requires location (-l) argument.")
		} else if flag.Arg(0) == "help" {
			usage()
		}
		action = flag.Arg(0)
	} else {
		usage()
	}

	client, err := clcv2.NewClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	/* If the first argument decodes as a hex value, assume it is a Hardware Group UUID */
	if _, err := hex.DecodeString(where); err == nil {
		server_action = false
	} else if utils.LooksLikeServerName(where) {
		server_action = true
		if *location != "" {
			fmt.Fprintf(os.Stderr, "WARNING: location (%s) ignored for %s\n", *location, where)
		}
	} else if *location != "" && where != "" {
		if group, err := client.GetGroupByName(where, *location); err != nil {
			exit.Errorf("Failed to resolve group name %q: %s", where, err)
		} else if group == nil {
			exit.Errorf("No group named %q was found on %s", where, *location)
		} else {
			where = group.Id
		}
	} else {
		server_action = true
	}

	if server_action { /* Server Action */
		switch action {
		case "show":
			showServer(client, where)
			os.Exit(0)
		}
		var server_actions = map[string]func(string) (string, error){
			"on":       client.PowerOnServer,
			"off":      client.PowerOffServer,
			"pause":    client.PauseServer,
			"reset":    client.ResetServer,
			"reboot":   client.RebootServer,
			"shutdown": client.ShutdownServer,
			"archive":  client.ArchiveServer,
			"delete":   client.DeleteServer,
			"snapshot": client.SnapshotServer,
		}

		/* Long-running commands that return a RequestID */
		handler, ok := server_actions[action]
		if !ok {
			exit.Fatalf("Unsupported action %s", action)
		}

		reqId, err = handler(where)
		if err != nil {
			exit.Fatalf("Server command %q failed: %s", action, err)
		}

	} else { /* Group Action */
		switch action {
		// TODO:  archive
		case "show":
			showGroup(client, where, *location)
			os.Exit(0)
		case "archive":
			reqId, err = client.ArchiveGroup(where)
		case "delete":
			reqId, err = client.DeleteGroup(where)
		default:
			exit.Errorf("Unsupported group action %q", action)
		}
		if err != nil {
			exit.Fatalf("Group command %q failed: %s", action, err)
		}
	}

	fmt.Printf("Request ID for %q action: %s\n", action, reqId)

	/* XXX
	locationStr := *location
	if server_action {
		locationStr = utils.ExtractLocationFromServerName(where)
	}
	client.PollDeploymentStatus(reqId, locationStr, *acctAlias, 1)
	*/
}

// Show server details
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
		"Group", "Description", "OS",
		"CPU", "Storage",
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
		grp.Name, server.Description, server.OsType,
		fmt.Sprint(server.Details.Cpu), fmt.Sprintf("%d GB", server.Details.StorageGb),
		strings.Join(IPs, " "),
		server.Details.PowerState, modifiedStr,
	})
	table.Render()

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

// Show group details
// @client:    authenticated CLCv2 Client
// @uuid:      hardware group UUID to use
// @location:  data centre location (needed to resolve @uuid)
func showGroup(client *clcv2.Client, uuid, location string) {
	if location == "" {
		exit.Errorf("Location is required in order to show the group hierarchy starting at %s", uuid)
	}

	root, err := client.GetGroups(location)
	if err != nil {
		exit.Fatalf("Failed to look up groups at %s: %s", location, err)
	}
	start := &root
	if uuid != "" {
		start = clcv2.FindGroupNode(start, func(g *clcv2.Group) bool {
			return g.Id == uuid
		})
		if start == nil {
			exit.Fatalf("Failed to look up UUID %s at %s", uuid, location)
		}
	}
	clcv2.PrintGroupHierarchy(start, "")
}
