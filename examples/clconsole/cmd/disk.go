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

// Flags
var addDiskFlags struct {
	Mount string // optional mountpoint for disk (changes disk type to 'partitioned')
}

func init() {
	var manageDisks = &cobra.Command{ // Top-level disk command
		Use:   "disk",
		Short: "Manage server disks",
		Long:  "Add, remove, or resize server disks",
	}

	diskAdd.Flags().StringVar(&addDiskFlags.Mount, "mount", "", "Optional mountpoint (otherwise disk type defaults to 'raw')")
	manageDisks.AddCommand(diskList, diskAdd, diskGrow, diskRemove)
	Root.AddCommand(manageDisks)
}

var diskAdd = &cobra.Command{
	Use:     "add  <server> <sizeGB>",
	Aliases: []string{"new"},
	Short:   "Add disk to server",
	Long:    "Adds a @sizeGB disk as 'raw' storage to @server",
	Example: "disk add WA1GRRT-RH5-05 2",
	PreRunE: checkArgs(2, "Need a server name and a disk size in GB"),
	RunE: func(cmd *cobra.Command, args []string) error {
		diskGB, err := strconv.ParseUint(args[1], 10, 32)
		if err != nil {
			return errors.Errorf("invalid disk-size value %q", args[1])
		}

		// First get the list of existing disks
		log.Printf("Getting %s details ...", args[0])
		server, err := client.GetServer(args[0])
		if err != nil {
			exit.Errorf("Failed to list details of server %q: %s", args[0], err)
		} else if len(server.Details.Snapshots) > 0 {
			return errors.Errorf("Unable to add disks since %s has a snapshot.", args[0])
		}

		disks := make([]clcv2.ServerAdditionalDisk, len(server.Details.Disks))
		for i := range server.Details.Disks {
			disks[i] = clcv2.ServerAdditionalDisk{
				Id:     server.Details.Disks[i].Id,
				SizeGB: server.Details.Disks[i].SizeGB,
			}
		}

		// Add  new disk at end - if path is specified, type must be set to 'partitioned'.
		newDisk := clcv2.ServerAdditionalDisk{SizeGB: uint32(diskGB), Type: "raw"}
		if addDiskFlags.Mount != "" {
			newDisk.Path = addDiskFlags.Mount
			newDisk.Type = "partitioned"
		}

		reqID, err := client.ServerSetDisks(args[0], append(disks, newDisk))
		if err != nil {
			log.Fatalf("failed to update the disk configuration on %q: %s", args[0], err)
		}

		log.Printf("%s adding %s %d GB disk: %s", args[0], newDisk.Type, diskGB, reqID)
		client.PollStatusFn(reqID, intvl, func(s clcv2.QueueStatus) {
			log.Printf("%s adding %s %d GB disk: %s", args[0], newDisk.Type, diskGB, s)
		})
		return nil
	},
}

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

		log.Printf("Getting %s details ...", args[0])
		server, err := client.GetServer(args[0])
		if err != nil {
			log.Fatalf("failed to list details of server %q: %s", args[0], err)
		} else if len(server.Details.Snapshots) > 0 {
			return errors.Errorf("Unable to change disk since %s has a snapshot.", args[0])
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
			log.Fatalf("failed to update the disk configuration on %q: %s", args[0], err)
		}

		log.Printf("%s resizing disk %s to %d GB: %s", args[0], id, diskGB, reqID)
		client.PollStatusFn(reqID, intvl, func(s clcv2.QueueStatus) {
			log.Printf("%s resizing disk %s to %d GB: %s", args[0], id, diskGB, s)
		})
		return nil
	},
}

var diskList = &cobra.Command{
	Use:     "ls  <server|group> [<server|group> ...]",
	Aliases: []string{"list", "show"},
	Short:   "List server disks",
	Long:    "Shows a table of disks for each server, or server in the group",
	PreRunE: checkAtLeastArgs(1, "Need at least 1 server or group to query"),
	Example: "disk ls CA2GRRT-PROD-02\ndisk ls dr_vms/",
	Run: func(cmd *cobra.Command, args []string) {
		if servnames, err := extractServerNames(args); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: failed to extract server names: %s\n", err)
		} else {
			var wg sync.WaitGroup
			var servers = make(chan clcv2.Server)

			// Query all servers in parallel
			for _, serverName := range servnames {
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

			// Waiter function which closes the channel once above goroutine is done
			go func() {
				wg.Wait()
				close(servers)
			}()

			// Consume the produced server details
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

var diskRemove = &cobra.Command{
	Use:     "rm  <server> <diskId> [<diskId> ... ]",
	Aliases: []string{"-", "remove", "del"},
	Short:   "Remove server disk(s)",
	Long:    "Remove one or more server disk(s)",
	Example: "disk del WA1GRRT-RH5-05 0:3 0:4 0:5\ndisk rm  WA1GRRT-RH5-05   3   4   5",
	PreRunE: checkAtLeastArgs(2, "Need a server name and at least 1 disk-ID"),
	RunE: func(cmd *cobra.Command, args []string) error {
		var ids clcv2.DiskIDList

		for _, idStr := range args[1:] {
			if id, err := clcv2.DiskIDFromString(idStr); err != nil {
				return errors.Errorf("Invalid disk ID %q", idStr)
			} else {
				ids.Add(id)
			}
		}

		log.Printf("Getting %s details ...", args[0])
		server, err := client.GetServer(args[0])
		if err != nil {
			log.Fatalf("failed to list details of server %q: %s", args[0], err)
		} else if len(server.Details.Snapshots) > 0 {
			return errors.Errorf("Unable to delete disks since %s has a snapshot.", args[0])
		}

		disks := make([]clcv2.ServerAdditionalDisk, 0)
		for i := range server.Details.Disks {
			if ids.Contains(server.Details.Disks[i].Id) {
				log.Printf("Will remove %s disk %s (%d GB)", args[0],
					server.Details.Disks[i].Id, server.Details.Disks[i].SizeGB)
			} else {
				disks = append(disks, clcv2.ServerAdditionalDisk{
					Id:     server.Details.Disks[i].Id,
					SizeGB: server.Details.Disks[i].SizeGB,
				})
			}
		}

		reqID, err := client.ServerSetDisks(args[0], disks)
		if err != nil {
			log.Fatalf("failed to update the disk configuration on %q: %s", args[0], err)
		}

		log.Printf("%s deleting disk %s: %s", args[0], ids, reqID)
		client.PollStatusFn(reqID, intvl, func(s clcv2.QueueStatus) {
			log.Printf("%s deleting disk %s: %s", args[0], ids, s)
		})
		return nil
	},
}
