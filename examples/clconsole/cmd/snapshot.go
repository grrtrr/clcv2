package cmd

/*
 * Snapshot operations
 */
import (
	"log"

	"github.com/grrtrr/clcv2"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// snapCreateFlags contains flags pertaining to creating a new snapshot
var snapCreateFlags struct {
	days int // Number of days to keep the snapshot for
}

func init() {
	// Create snapshot
	var createSnapshot = &cobra.Command{
		Use:     "snapshot  [group|server [group|server]...]",
		Aliases: []string{"snap", "sn"},
		Short:   "Snapshot server(s)",
		Long:    "Create new server snapshot, replacing previous snapshot if it exists",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.Errorf("Need at least 1 server to snapshot")
			}
			if snapCreateFlags.days <= 0 || snapCreateFlags.days > 10 {
				return errors.Errorf("Invalid number of days %d - must be in the range 1..10", snapCreateFlags.days)
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			serverCmd("create new snapshot", snapshotHandler, args)
		}}
	createSnapshot.Flags().IntVar(&snapCreateFlags.days, "days", 10, "Number of days to keep the snapshot for")
	Root.AddCommand(createSnapshot)

	// Delete snapshot
	Root.AddCommand(&cobra.Command{
		// FIXME: should be sub-command of snapshot
		Use:     "delsnap  server [server...]",
		Aliases: []string{"rmsnap"},
		Short:   "Delete snapshot of server(s)",
		Long:    "Delete server snapshot if it exists (error condition of no snapshot exists)",
		PreRunE: checkAtLeastArgs(1, "Need at least 1 server to remove snapshots from"),
		Run: func(cmd *cobra.Command, args []string) {
			serverCmd("delete snapshot", client.DeleteSnapshot, args)
		}})

	// Revert to snapshot
	Root.AddCommand(&cobra.Command{
		Use:     "revert  server [server...]",
		Aliases: []string{"rev"},
		Short:   "Revert server(s) to snapshot",
		Long:    "Revert server(s) to last snapshot (error condition if no snapshot exists)",
		PreRunE: checkAtLeastArgs(1, "Need at least 1 server to revert"),
		Run: func(cmd *cobra.Command, args []string) {
			serverCmd("revert to snapshot", client.RevertToSnapshot, args)
		}})
}

// snapshotHandler is a convenience function to automatically delete an existing snapshot before creating a new one.
// This is to satisfy the use case "I want to snapshot what I just did", without having to run multiple commands.
// Note: relies on 'client' and 'intvl' global variables.
func snapshotHandler(serverId string) (reqId string, err error) {
	if reqID, err := client.DeleteSnapshot(serverId); err == nil {
		log.Printf("%s delete current snapshot: %s", serverId, reqID)
		client.PollStatusFn(reqID, intvl, func(s clcv2.QueueStatus) {
			log.Printf("%s delete current snapshot: %s", serverId, s)
		})
	} else if err != clcv2.ErrNoSnapshot {
		return "", errors.Errorf("%s: failed to delete potentially existing snapshot: %s", serverId, err)
	}
	return client.CreateSnapshot(serverId, snapCreateFlags.days)
}
