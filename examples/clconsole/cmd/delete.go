package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/grrtrr/clcv2"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

// deleteFlags determine how to perform deletions
var deleteFlags struct {
	recurse bool
	keep    bool
}

func init() {
	delete.Flags().BoolVarP(&deleteFlags.recurse, "recurse", "r", true, "When deleting a group directory, also delete all of its sub-directories")
	delete.Flags().BoolVarP(&deleteFlags.keep, "keep-directory", "k", false, "Keep any specified group directories (only delete their contents)")

	Root.AddCommand(delete)
}

var delete = &cobra.Command{
	Use:     "rm  [group|server [group|server]...]",
	Aliases: []string{"remove", "del", "delete", "clean-up", "rd"},
	Short:   "Delete server(s)/group(s) (CAUTION)",
	Long:    "Completely and irreversibly removes servers/groups - USE WITH CAUTION",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.Errorf("Need at least 1 server or group to remove")
		}
		for _, arg := range args {
			if arg == "" {
				return errors.Errorf("%s: you requested %q, which means EVERYTHING IN %s - refusing to continue",
					cmd.Name(), arg, strings.ToUpper(conf.Location))
			}
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var root, g *clcv2.Group
		var eg = new(errgroup.Group)

		groups, servers, err := resolveNames(args)
		if err != nil {
			return errors.Errorf("%s: %s", cmd.Name(), err)
		}

		if len(groups) > 0 {
			if conf.Location == "" {
				return errors.Errorf("Location argument (-l) is required in order to traverse nested groups.")
			} else if root, err = client.GetGroups(conf.Location); err != nil {
				log.Fatalf("Failed to query group structure in %s: %s", conf.Location, err)
			}
		}

		for _, srv := range servers {
			deleteSingleServer(eg, srv)
		}

		for _, grp := range groups {
			grp := grp
			eg.Go(func() error {
				if grp == "" {
					return errors.Errorf("Not accepting %q, as it means to delete everything in %s", grp, conf.Location)
				} else if g = clcv2.FindGroupNode(root, func(g *clcv2.Group) bool { return g.Id == grp }); g == nil {
					return errors.Errorf("Failed to look up group %q in %s - is the location correct?", grp, conf.Location)
				}

				groupDir, err := clcv2.WalkGroupHierarchy(context.TODO(), g, nil) // nil callback here, so will process fast
				if err != nil {
					return errors.Errorf("Failed to process %s group hierarchy in %s: %s", grp, conf.Location, err)
				}

				if !deleteFlags.keep && deleteFlags.recurse { // wipe this directory and all of its children
					deleteSingleGroup(eg, groupDir)
				} else if len(groupDir.Servers) == 0 && (!deleteFlags.recurse || len(groupDir.Groups) == 0) {
					if l := len(groupDir.Groups); l == 0 {
						log.Printf("Nothing to delete in %s - directory empty", groupDir.Name)
					} else if !deleteFlags.recurse {
						log.Printf("Nothing to delete in %s (will not delete %d subdirectories since --recurse=false)",
							groupDir.Name, l)
					}
				} else {
					for _, srv := range groupDir.Servers {
						deleteSingleServer(eg, srv)
					}

					if deleteFlags.recurse {
						for _, grp := range groupDir.Groups {
							deleteSingleGroup(eg, grp)
						}
					}
				}
				return nil
			})
		}
		if err = eg.Wait(); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		}
		return nil
	},
}

// deleteSingleServer is a helper function to delete server @srv within error group @eg
func deleteSingleServer(eg *errgroup.Group, srv string) {
	eg.Go(func() error {
		if reqID, err := client.DeleteServer(srv); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR deleting server %s: %s\n", srv, err)
		} else if reqID != "" {
			log.Printf("Deleting %s: %s", srv, reqID)
			client.PollStatusFn(reqID, intvl, func(s clcv2.QueueStatus) {
				log.Printf("Deleting %s: %s", srv, s)
			})
		} else {
			fmt.Println("schase")
		}
		// For single items we do not feed back the error to the error group, just log it.
		return nil
	})
}

// deleteSingleGroup is analogous to deleteSingleServer group, deleting a group @g with all of its children
func deleteSingleGroup(eg *errgroup.Group, grp *clcv2.GroupInfo) {
	eg.Go(func() error {
		if reqID, err := client.DeleteGroup(grp.ID); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR deleting group %s (%s): %s\n", grp.Name, grp.ID, err)
		} else if reqID != "" {
			log.Printf("Deleting %s (%s): %s", grp.Name, grp.ID, reqID)
			client.PollStatusFn(reqID, intvl, func(s clcv2.QueueStatus) {
				log.Printf("Deleting %s (%s): %s", grp.Name, grp.ID, s)
			})
		}
		// For single items we do not feed back the error to the error group, just log it.
		return nil
	})
}
