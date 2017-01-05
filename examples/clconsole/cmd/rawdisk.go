package cmd

import (
	"log"
	"strconv"

	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var Rawdisk = &cobra.Command{
	Use:     "rawdisk <server> <sizeGB>",
	Aliases: []string{"storage", "disk"},
	Short:   "Add storage to server",
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
			return errors.Errorf("Failed to list details of server %q: %s", args[0], err)
		}

		diskGB, err := strconv.ParseUint(args[1], 10, 32)
		if err != nil {
			return errors.Errorf("Invalid disk-size value %q", args[1])
		}

		disks := make([]clcv2.ServerAdditionalDisk, len(server.Details.Disks))
		for i := range server.Details.Disks {
			disks[i] = clcv2.ServerAdditionalDisk{
				Id:     server.Details.Disks[i].Id,
				SizeGB: server.Details.Disks[i].SizeGB,
			}
		}

		reqID, err := client.ServerSetDisks(args[0], append(disks,
			clcv2.ServerAdditionalDisk{SizeGB: uint32(diskGB), Type: "raw"}))
		if err != nil {
			exit.Fatalf("failed to update the disk configuration on %q: %s", args[0], err)
		}

		if reqID != "" {
			client.PollStatusFn(reqID, intvl, func(s clcv2.QueueStatus) {
				log.Printf("Adding %s GB to %s: %s", args[1], args[0], s)
			})
		}
		return nil
	},
}

func init() {
	Root.AddCommand(Rawdisk)
}
