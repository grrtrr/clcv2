package cmd

/*
 * Snapshot operations
 */
import (
	"github.com/spf13/cobra"
)

func init() {
	Root.AddCommand(&cobra.Command{
		Use:     "snapshot  [group|server [group|server]...]",
		Aliases: []string{"snap", "sn"},
		Short:   "Snapshot server(s)",
		Long:    "Create new server snapshot, replacing older snapshots if they exist",
		Run: func(cmd *cobra.Command, args []string) {
			serverCmd("snapshot", client.SnapshotServer, args)
		}})

	Root.AddCommand(&cobra.Command{
		Use:     "delsnap  [group|server [group|server]...]",
		Aliases: []string{"rmsnap", "delete-snapshot"},
		Short:   "Delete snapshot of server(s)",
		Long:    "Delete server snapshot if it exists (error condition of no snapshot exists)",
		Run: func(cmd *cobra.Command, args []string) {
			serverCmd("delete snapshot", client.DeleteSnapshot, args)
		}})

	Root.AddCommand(&cobra.Command{
		Use:     "revert  [group|server [group|server]...]",
		Aliases: []string{"rev"},
		Short:   "Revert server(s) to snapshot",
		Long:    "Revert server(s) to last snapshot (error condition if no snapshot exists)",
		Run: func(cmd *cobra.Command, args []string) {
			serverCmd("revert to snapshot", client.RevertToSnapshot, args)
		}})
}
