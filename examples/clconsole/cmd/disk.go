package cmd

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var diskGrow = &cobra.Command{
	Use:     "grow  <server> <diskID> <sizeGB>",
	Aliases: []string{"resize", "increase"},
	Short:   "Resize server disk",
	Long:    "Resize disk @diskID of @server to @sizeGB (disk ID uses [<maj>:]<min> format)",
	Example: "grow   CA2GRRT-PROD-02 0:3 256\nresize CA2GRRT-PROD-02   3 256",
	PreRunE: checkArgs(3, "Need a server, a disk ID, and the new disk size in GB"),
	RunE: func(cmd *cobra.Command, args []string) error {
		var found bool

		id, err := clcv2.DiskIDFromString(args[1])
		if err != nil {
			return errors.Errorf("invalid disk ID %q", args[1])
		}

		diskGB, err := strconv.ParseUint(args[2], 10, 32)
		if err != nil {
			return errors.Errorf("invalid size/GB value %q", args[2])
		}

		server, err := client.GetServer(args[0])
		if err != nil {
			exit.Fatalf("failed to list details of server %q: %s", args[0], err)
		}

		disks := make([]clcv2.ServerAdditionalDisk, len(server.Details.Disks))
		for i := range server.Details.Disks {
			disks[i] = clcv2.ServerAdditionalDisk{
				Id:     server.Details.Disks[i].Id,
				SizeGB: server.Details.Disks[i].SizeGB,
			}
			if disks[i].Id == id {
				found = true
				// The API does not allow to reduce the size of an existing disk.
				if uint32(diskGB) <= disks[i].SizeGB {
					return errors.Errorf("%s disk %s is already at %d GB.", args[0], id, disks[i].SizeGB)
				}
				log.Printf("Resizing %s disk %s from %d to %d GB ...", args[0], id, disks[i].SizeGB, diskGB)
				disks[i].SizeGB = uint32(diskGB)
			}
		}

		// Make sure the disk exists: otherwise the API will return an empty 204 response - and no status link.
		if !found {
			return errors.Errorf("%s does not have a disk with ID %s", args[0], id)
		}

		reqID, err := client.ServerSetDisks(args[0], disks)
		if err != nil {
			exit.Fatalf("failed to update the disk configuration on %q: %s", args[0], err)
		}

		log.Printf("%s resizing disk %s to %d GB: %s", args[0], id, diskGB, reqID)
		client.PollStatusFn(reqID, intvl, func(s clcv2.QueueStatus) {
			log.Printf("%s resizing disk %s to %d GB: %s", args[0], id, diskGB, s)
		})
		return nil
	},
}

var diskRemove = &cobra.Command{
	Use:     "rm  <server> <disk-ID>",
	Aliases: []string{"remove", "del"},
	Short:   "Remove server disk",
	Long:    "Remove single server disk", // XXX
	// Example: // XXX
	PreRunE: checkArgs(2, "Need a server name and a disk ID"),
	Run: func(cmd *cobra.Command, args []string) {
		// FIXME
		fmt.Fprintf(os.Stderr, "FIXME: this is not implemented yet")
	},
}

func init() {
	var manageDisks = &cobra.Command{ // Top-level disk command
		Use:   "disk",
		Short: "Manage server disks",
		Long:  "Add, remove, or grow server disks",
	}

	manageDisks.AddCommand(diskList, diskAdd, diskGrow, diskRemove)
	Root.AddCommand(manageDisks)
}

var diskAdd = &cobra.Command{
	Use:     "new <server> <sizeGB>",
	Aliases: []string{"add", "raw", "rawdisk"},
	Short:   "Add disk to server",
	Long:    "Adds @sizeGB of storage as 'raw' disk to @server",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return errors.Errorf("Need a server name and a disk size in GB")
		} else if _, err := strconv.ParseUint(args[1], 10, 32); err != nil {
			return errors.Errorf("Invalid disk-size value %q", args[1])
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// First get the list of existing disks
		server, err := client.GetServer(args[0])
		if err != nil {
			exit.Errorf("Failed to list details of server %q: %s", args[0], err)
		}

		diskGB, err := strconv.ParseUint(args[1], 10, 32)
		if err != nil {
			return errors.Errorf("invalid disk-size value %q", args[1])
		}

		disks := make([]clcv2.ServerAdditionalDisk, len(server.Details.Disks))
		for i := range server.Details.Disks {
			disks[i] = clcv2.ServerAdditionalDisk{
				Id:     server.Details.Disks[i].Id,
				SizeGB: server.Details.Disks[i].SizeGB,
			}
		}

		// Add new disk at the end of the list of existing disks.
		reqID, err := client.ServerSetDisks(args[0], append(disks,
			clcv2.ServerAdditionalDisk{SizeGB: uint32(diskGB), Type: "raw"}))
		if err != nil {
			exit.Fatalf("failed to update the disk configuration on %q: %s", args[0], err)
		}

		log.Printf("%s adding %d GB raw storage: %s", args[0], diskGB, reqID)
		client.PollStatusFn(reqID, intvl, func(s clcv2.QueueStatus) {
			log.Printf("%s adding %d GB: %s", args[0], diskGB, s)
		})
		return nil
	},
}

var diskList = &cobra.Command{
	Use:     "ls <server>",
	Aliases: []string{"list"},
	Short:   "List server disks",
	Long:    "Shows a tabulated breakdown of the disks of each server",
	PreRunE: checkArgs(1, "Need a server name to query"),
	Run: func(cmd *cobra.Command, args []string) {
		if servnames, err := extractServerNames(args); err != nil { // just show a list of servers
			fmt.Fprintf(os.Stderr, "ERROR: failed to extract server names: %s\n", err)
		} else {
			var wg sync.WaitGroup
			var servers = make(chan clcv2.Server)

			for _, serverName := range servnames { // Query all servers in parallel
				wg.Add(1)
				go func(s string) {
					if server, err := client.GetServer(s); err != nil {
						fmt.Fprintf(os.Stderr, "ERROR: failed to get % details: %s\n", s, err)
					} else {
						servers <- server
					}
					wg.Done()
				}(serverName)
			}

			go func() { // Waiter function which closes the channel once we are done
				wg.Wait()
				close(servers)
			}()

			for server := range servers {
				fmt.Printf("%s disks (total %d GB):\n", server.Name, server.Details.StorageGb)
				table := tablewriter.NewWriter(os.Stdout)
				table.SetAutoFormatHeaders(false)
				table.SetAlignment(tablewriter.ALIGN_RIGHT)
				table.SetAutoWrapText(true)

				table.SetHeader([]string{"ID", "Size/GB", "Paths"})
				for _, d := range server.Details.Disks {
					table.Append([]string{string(d.Id), fmt.Sprint(d.SizeGB), strings.Join(d.PartitionPaths, ", ")})
				}
				table.Render()
				fmt.Printf("\n")
			}
		}
	},
}
