package cmd

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"

	humanize "github.com/dustin/go-humanize"
	"github.com/grrtrr/clcv2"
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
	Show.Flags().BoolVar(&showGroupTree, "tree", false, "Display nested group structure in tree format")
	Show.Flags().BoolVar(&showIP, "ips", false, "Print group structure with server IPs (implies --group and --tree)")

	Root.AddCommand(Show)
}

var Show = &cobra.Command{
	Use:     "ls  [group|server [group|server]...]",
	Aliases: []string{"dir", "show"},
	Short:   "Show server(s)/groups(s)",
	Long:    "Display detailed server/group information. Group information requires -l to be set.",
	RunE: func(cmd *cobra.Command, args []string) error {
		var nodeCallback func(context.Context, *clcv2.GroupInfo) error
		var servers, groups []string
		var root *clcv2.Group
		var err error

		// Showing IP information implies printing the nested group structure
		if showIP {
			showGroupTree = true
			showGroupDetails = true
			nodeCallback = queryServerState
		}

		switch l := len(args); l {
		case 1:
			// Allow user to specify data center name as only argument
			if regexp.MustCompile(`^[[:alpha:]]{2}\d$`).MatchString(args[0]) {
				conf.Location = args[0]
				args = append(args[:0], "")
				showGroupTree = true
			}
		case 0:
			// The default behaviour is to list all the servers/groups in the default data centre.
			args = append(args, "")
			showGroupTree = true
		}

		if showGroupDetails || showGroupTree {
			for _, name := range args {
				isServer, where, err := groupOrServer(name)
				if err != nil {
					fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
				} else if isServer {
					servers = append(servers, where)
				} else if conf.Location == "" && showGroupTree {
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
				root, err = client.GetGroups(conf.Location)
				if err != nil {
					return errors.Errorf("Failed to look up groups at %s: %s", conf.Location, err)
				}
			}

			if showGroupTree {
				start := root
				if uuid != "" {
					start = clcv2.FindGroupNode(root, func(g *clcv2.Group) bool { return g.Id == uuid })
					if start == nil {
						return errors.Errorf("Failed to look up group %q in %s - is the location correct?", uuid, conf.Location)
					}
				}
				tree, err := clcv2.WalkGroupHierarchy(context.TODO(), start, nodeCallback)
				if err != nil {
					return errors.Errorf("failed to process %s group hierarchy: %s", conf.Location, err)
				}
				printGroupStructure(tree, "")
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

// Pretty-printer for traversal of nested group structure.
func printGroupStructure(g *clcv2.GroupInfo, indent string) {
	var groupLine string

	if g.Type != "default" { // 'Archive' or similar: make it stand out
		groupLine = fmt.Sprintf("%s[%s]/", indent, g.Name)
	} else {
		groupLine = fmt.Sprintf("%s%s/", indent, g.Name)
	}

	fmt.Printf("%-70s %s\n", groupLine, g.ID)

	for _, s := range g.Servers {
		fmt.Printf("%s%s\n", indent+"    ", s)
	}

	//	sort.Sort(g.Groups)
	for _, g := range g.Groups {
		printGroupStructure(g, indent+"    ")
	}
}

// queryServerState processes a single clcv2.GroupInfo node in isolation, adding server information
// NOTE: requires 'client' variable to be in enclosing scope
func queryServerState(ctx context.Context, node *clcv2.GroupInfo) error {
	var serverEntries = make(chan string)
	var g, gctx = errgroup.WithContext(ctx)

	for _, id := range node.Servers {
		id := id
		g.Go(func() error {
			srv, err := client.GetServer(id)
			if err != nil {
				return errors.Errorf("failed to get %q server information: %s", id, err)
			}

			servLine := id
			if len(srv.Details.Snapshots) > 0 { // add a tilde to indicate it has a snapshot
				servLine += "~"
			}

			if srv.Details.PowerState == "started" { // add an asterisk to indicate it's on
				servLine += "*"
			}

			select {
			case serverEntries <- fmt.Sprintf("%-50s %s", servLine, strings.Join(srv.IPs(), ", ")):
			case <-gctx.Done():
				return gctx.Err()
			}
			return nil
		})
	}

	go func() {
		g.Wait()
		close(serverEntries)
	}()

	node.Servers = node.Servers[:0]
	for srv := range serverEntries {
		node.Servers = append(node.Servers, srv)
	}
	return g.Wait()
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
		table.SetAlignment(tablewriter.ALIGN_CENTER)
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
