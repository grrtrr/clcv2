package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/grrtrr/clcv2"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

func init() {
	Root.AddCommand(&cobra.Command{
		Use:     "on  [group|server [group|server]...]",
		Aliases: []string{"start", "power-on", "up"},
		Short:   "Power on server(s)",
		Long:    "Powers on server(s), or resume from paused state",
		Run: func(cmd *cobra.Command, args []string) {
			serverCmd("power-on", client.PowerOnServer, args)
		}})

	Root.AddCommand(&cobra.Command{
		Use:     "off  [group|server [group|server]...]",
		Aliases: []string{"power-off"},
		Short:   "Power-off server(s)",
		Long:    "Do a forceful power-off of server(s) (as opposed to a soft OS-level shutdown)",
		Run: func(cmd *cobra.Command, args []string) {
			serverCmd("power-off", client.PowerOffServer, args)
		}})

	Root.AddCommand(&cobra.Command{
		Use:     "pause  [group|server [group|server]...]",
		Aliases: []string{"suspend"},
		Short:   "Pause server(s)",
		Long:    "Pause (suspend) server(s); can be resumed via 'on'",
		Run: func(cmd *cobra.Command, args []string) {
			serverCmd("pause", client.PauseServer, args)
		}})

	Root.AddCommand(&cobra.Command{
		Use:   "reset  [group|server [group|server]...]",
		Short: "Reset server(s)",
		Long:  "Performs hard/forced power-cycle, like pressing the physical 'reset' button",
		Run: func(cmd *cobra.Command, args []string) {
			serverCmd("reset", client.ResetServer, args)
		}})

	Root.AddCommand(&cobra.Command{
		Use:   "reboot  [group|server [group|server]...]",
		Short: "Reboot server(s)",
		Long:  "Soft (OS-level) reboot",
		Run: func(cmd *cobra.Command, args []string) {
			serverCmd("reboot", client.RebootServer, args)
		}})

	Root.AddCommand(&cobra.Command{
		Use:     "shutdown  [group|server [group|server]...]",
		Aliases: []string{"stop"},
		Short:   "Shutdown server(s)",
		Long:    "Soft (OS-level) shutdown, followed by power-off",
		Run: func(cmd *cobra.Command, args []string) {
			serverCmd("shutdown", client.ShutdownServer, args)
		}})

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
		Aliases: []string{"delete-snapshot"},
		Short:   "Delete snapshot of server(s)",
		Long:    "Delete server snapshot if it exists (error condition of no snapshot exists)",
		Run: func(cmd *cobra.Command, args []string) {
			serverCmd("delete snapshot", client.DeleteSnapshot, args)
		}})

	Root.AddCommand(&cobra.Command{
		Use:   "revert  [group|server [group|server]...]",
		Short: "Revert server(s) to snapshot",
		Long:  "Revert server(s) to last snapshot (error condition if no snapshot exists)",
		Run: func(cmd *cobra.Command, args []string) {
			serverCmd("revert to snapshot", client.RevertToSnapshot, args)
		}})
}

// serverCmd wraps common server tasks
// @action: name of the command
// @hdlr:   server action, taking a server ID as argument and returning the status ID, or an error
// @args:   command arguments (server or group names) to loop over
func serverCmd(action string, hdlr func(string) (string, error), args []string) error {
	var eg errgroup.Group

	servers, err := extractServerNames(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return err
	}

	for _, name := range servers {
		name := name
		eg.Go(func() error {
			reqID, err := hdlr(name)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR %s %s: %s\n", name, action, err)
			} else {
				log.Printf("%s %s: %s", name, action, reqID)

				client.PollStatusFn(reqID, intvl, func(s clcv2.QueueStatus) {
					log.Printf("%s %s: %s", name, action, s)
				})
			}
			return err
		})
	}
	return eg.Wait()
}

// extractServerNames extracts all server names specified via @args
// - server names contained in @args are returned directly
// - for group names, it recursively collects all server names contained in the group
func extractServerNames(args []string) (ret []string, err error) {
	var root *clcv2.Group
	var groupVisitor func(g *clcv2.Group)

	groupVisitor = func(g *clcv2.Group) {
		for _, l := range g.Links {
			if l.Rel == "server" {
				ret = append(ret, l.Id)
			}
		}
		for idx := range g.Groups {
			groupVisitor(&g.Groups[idx])
		}
	}

	for _, name := range args {
		if isServer, where, err := groupOrServer(name); err != nil {
			return nil, err
		} else if isServer {
			ret = append(ret, where)
		} else if location == "" {
			return nil, errors.Errorf("Location argument (-l) is required in order to traverse group %s", name)
		} else if root, err = client.GetGroups(location); err != nil {
			return nil, errors.Errorf("Failed to look up groups at %s: %s", location, err)
		} else {
			start := root
			if where != "" {
				start = clcv2.FindGroupNode(root, func(g *clcv2.Group) bool { return g.Id == where })
				if start == nil {
					return nil, errors.Errorf("Failed to look up group %q in %s - is the location correct?", where, location)
				}
			}
			groupVisitor(start)
		}
	}
	return ret, nil
}
