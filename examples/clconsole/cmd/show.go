package cmd

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"sync"

	humanize "github.com/dustin/go-humanize"
	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/clcv2/utils"
	"github.com/grrtrr/exit"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// Flags
var (
	showGroupDetails bool // whether to print group details instead of showing the contained servers
	showGroupTree    bool // whether to display groups in tree format
	showIP           bool // whether to just display server IPs (implies showGroupTree and showGroupDetails)
)

func init() {
	Show.Flags().BoolVar(&showGroupDetails, "group", false, "Print group details rather than the contained servers")
	Show.Flags().BoolVar(&showGroupTree, "tree", true, "Display nested group structure in tree format")
	Show.Flags().BoolVar(&showIP, "ips", false, "Print group structure with server IPs (implies --group and --tree)")

	Root.AddCommand(Show)
}

var Show = &cobra.Command{
	Use:   "show  [group|server [group|server]...]",
	Short: "Show server(s)/groups(s)",
	Long:  "Display detailed server/group information. Group information requires -l to be set.",
	RunE: func(cmd *cobra.Command, args []string) error {
		var servers, groups []string
		var groupPrinter = printGroupStructure
		var root *clcv2.Group
		var err error

		// Showing IP information implies printing the nested group structure
		if showIP {
			showGroupTree = true
			showGroupDetails = true
			groupPrinter = printGroupWithServerIPs
		}

		// The default behaviour is to list all the servers/groups in the default data centre.
		if len(args) == 0 {
			args = append(args, "")
			showGroupDetails = true
		}

		if showGroupDetails || showGroupTree {
			for _, name := range args {
				isServer, where, err := groupOrServer(name)
				if err != nil {
					fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
				} else if isServer {
					servers = append(servers, where)
				} else if location == "" && showGroupTree {
					// Printing group trees: requires to resolve the root first.
					return errors.Errorf("Location argument (-l) is required in order to traverse nested groups.")
				} else {
					groups = append(groups, where)
				}
			}
		} else if servers, err = extractServerNames(args); err != nil { // just show a list of servers
			fmt.Fprintf(os.Stderr, "Failed to extract server names: %s\n", err)
			return nil
		}

		// Aggregate displaying of servers
		if l := len(servers); l == 1 {
			showServerByName(client, servers[0])
		} else if l > 1 {
			showServers(client, servers)
		}

		for _, uuid := range groups {
			if (uuid == "" || showGroupTree) && root == nil {
				root, err = client.GetGroups(location)
				if err != nil {
					return errors.Errorf("Failed to look up groups at %s: %s", location, err)
				}
			}

			if showGroupTree {
				start := root
				if uuid != "" {
					start = clcv2.FindGroupNode(root, func(g *clcv2.Group) bool { return g.Id == uuid })
					if start == nil {
						return errors.Errorf("Failed to look up UUID %s in %s - is this the correct value?", uuid, location)
					}
				}
				clcv2.VisitGroupHierarchy(start, groupPrinter, "")
			} else if uuid == "" {
				showGroup(client, root)
			} else if rootNode, err := client.GetGroup(uuid); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to query HW group %q: %s\n", uuid, err)
			} else {
				showGroup(client, rootNode)
			}
		}
		return nil
	},
}

// groupOrServer decides whether @name refers to a CLCv2 hardware group or a server.
// It indicates the result via a returned boolean flag, and resolves @where into @id.
func groupOrServer(name string) (isServer bool, id string, err error) {
	// Strip trailing slashes that hint at a group name (but are not part of the CLC name).
	if where := strings.TrimRight(name, "/"); where == "" {
		// An emtpy name by default refers to all entries in the default data centre.
		return false, "", nil
	} else if _, errHex := hex.DecodeString(where); errHex == nil {
		/* If the first argument decodes as a hex value, assume it is a Hardware Group UUID */
		return false, where, nil
	} else if utils.LooksLikeServerName(where) { /* Starts with a location identifier and is not hex ... */
		return true, strings.ToUpper(where), nil
	} else if location != "" { /* Fallback: assume it is a group */
		if group, err := client.GetGroupByName(where, location); err != nil {
			return false, where, errors.Errorf("failed to resolve group name %q: %s", where, err)
		} else if group == nil {
			return false, where, errors.Errorf("no group named %q was found in %s", where, location)
		} else {
			return false, group.Id, nil
		}
		err = errors.Errorf("unable to resolve group name %q in %s", where, location)
	} else if location == "" {
		err = errors.Errorf("%q looks like a group name - need a location (-l argument) to resolve it.", where)
	} else {
		err = errors.Errorf("unable to determine whether %q is a server or a group", where)
	}
	return
}

// showGroup displays details of Hardware Group folder @root
func showGroup(client *clcv2.CLIClient, root *clcv2.Group) {
	fmt.Printf("Group %q in %s:\n", root.Name, root.LocationId)
	fmt.Printf("ID:            %s\n", root.Id)
	fmt.Printf("Description:   %s\n", root.Description)
	fmt.Printf("Type:          %s\n", root.Type)
	fmt.Printf("Status:        %s\n", root.Status)

	if len(root.CustomFields) > 0 {
		fmt.Println("Custom fields:", root.CustomFields)
	}

	// ChangeInfo
	createdStr := humanize.Time(root.ChangeInfo.CreatedDate)
	/* The CreatedBy field can be an email address, or an API Key (hex string) */
	if _, err := hex.DecodeString(root.ChangeInfo.CreatedBy); err == nil {
		createdStr += " via API Key"
	} else {
		createdStr += " by " + root.ChangeInfo.CreatedBy
	}
	fmt.Printf("Created:       %s\n", createdStr)

	modifiedStr := humanize.Time(root.ChangeInfo.ModifiedDate)
	/* The ModifiedBy field can be an email address, or an API Key (hex string) */
	if _, err := hex.DecodeString(root.ChangeInfo.ModifiedBy); err == nil {
		modifiedStr += " via API Key"
	} else {
		modifiedStr += " by " + root.ChangeInfo.ModifiedBy
	}
	fmt.Printf("Modified:      %s\n", modifiedStr)

	// Servers
	fmt.Printf("#Servers:      %d\n", root.Serverscount)
	if root.Serverscount > 0 {
		var servers []string

		if sl := clcv2.ExtractLinks(root.Links, "server"); len(sl) > 0 {
			for _, s := range sl {
				servers = append(servers, s.Id)
			}
			fmt.Printf("Servers:       %s\n", strings.Join(servers, ", "))
		}
	}

	// Sub-groups
	if len(root.Groups) > 0 {
		fmt.Printf("\nGroups of %s:\n", root.Name)
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoFormatHeaders(false)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetAutoWrapText(true)

		table.SetHeader([]string{"Name", "UUID", "Description", "#Servers", "Type"})
		for _, g := range root.Groups {
			table.Append([]string{g.Name, g.Id, g.Description, fmt.Sprint(g.Serverscount), g.Type})
		}
		table.Render()
	} else {
		fmt.Printf("Sub-groups:    none\n")
	}
}

// clcv2.VisitGroupHierarchy callback function to print nested group structure
func printGroupStructure(g *clcv2.Group, arg interface{}) interface{} {
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

// clcv2.VisitGroupHierarchy callback function to print servers along with their IP addresses
// NOTE: requires 'client' variable to be in enclosing scope
func printGroupWithServerIPs(g *clcv2.Group, arg interface{}) interface{} {
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
				exit.Fatalf("failed to get IPs of %q in %s: %s", l.Id, g.Name, err)
			}

			servLine := fmt.Sprintf("%s%s", indent+"    ", l.Id)
			fmt.Printf("%-50s %s\n", servLine, strings.Join(ips, ", "))
		}
	}
	return indent + "    "
}

// Show details of a single server @name
// @client:    authenticated CLCv2 Client
// @servname:  server name
func showServerByName(client *clcv2.CLIClient, servname string) {
	if server, err := client.GetServer(servname); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to list details of server %q: %s\n", servname, err)
	} else {
		showServer(client, server)
	}
}

// Show details of a single server
func showServer(client *clcv2.CLIClient, server clcv2.Server) {
	grp, err := client.GetGroup(server.GroupId)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to resolve group UUID: %s", err)
		return
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

// Show condensed details of multiple servers
// @client:    authenticated CLCv2 Client
// @servnames: server names
func showServers(client *clcv2.CLIClient, servnames []string) {
	var wg sync.WaitGroup
	var truncate = func(s string, maxlen int) string {
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
		"Name", "Group", "Description", "OS",
		"IP", "CPU", "Mem", "Storage",
		"Status", "Last Change",
	})

	for _, servname := range servnames {
		servname := servname
		wg.Add(1)
		go func() {
			defer wg.Done()
			server, err := client.GetServer(servname)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to list details of server %q: %s", servname, err)
				return
			}

			grp, err := client.GetGroup(server.GroupId)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to resolve %s group UUID: %s\n", servname, err)
				return
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

			status := server.Details.PowerState
			if server.Details.InMaintenanceMode {
				status = "MAINTENANCE"
			} else if server.Status != "active" {
				status = server.Status
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

			// Append a tilde (~) to indicate it has snapshots
			serverName := server.Name
			if len(server.Details.Snapshots) > 0 {
				serverName += "~"
			}

			table.Append([]string{
				serverName, grp.Name, truncate(desc, 30), truncate(server.OsType, 15),
				strings.Join(IPs, " "),
				fmt.Sprint(server.Details.Cpu), fmt.Sprintf("%d G", server.Details.MemoryMb/1024),
				fmt.Sprintf("%d G", server.Details.StorageGb),
				status, modifiedStr,
			})
		}()
	}
	wg.Wait()
	table.Render()
}
