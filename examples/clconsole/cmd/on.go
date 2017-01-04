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

// extractServerNames extracts all server names specified via @args
// - server names contained in @args are returned directly
// - for group names, it recursively collects all server names contained in the group
func extractServerNames(args []string) (ret []string, err error) {
	var root *clcv2.Group
	var groupCallback = func(g *clcv2.Group, arg interface{}) interface{} {
		for _, l := range g.Links {
			if l.Rel == "server" {
				log.Printf("%s append %s", g.Name, l.Id)
				ret = append(ret, l.Id)
			}
		}
		return nil
	}

	for _, name := range args {
		if isServer, where, err := groupOrServer(name); isServer {
			ret = append(ret, where)
		} else if location == "" {
			return nil, errors.Errorf("Location argument (-l) is required in order to traverse group %s", name)
		} else if root, err = client.GetGroups(location); err != nil {
			return nil, errors.Errorf("Failed to look up groups at %s: %s\n", location, err)
		} else {
			start := root
			if where != "" {
				start = clcv2.FindGroupNode(root, func(g *clcv2.Group) bool { return g.Id == where })
				if start == nil {
					return nil, errors.Errorf("Failed to look up UUID %s in %s - is this the correct value?", where, location)
				}
			}
			clcv2.VisitGroupHierarchy(start, groupCallback, nil)
		}
	}
	return ret, nil
}

var On = &cobra.Command{
	Use:     "on [group|server [group|server]...]",
	Aliases: []string{"start", "up"},
	Short:   "Power on server(s) (or resume from paused state)",
	PreRun: func(cmd *cobra.Command, args []string) {

	},
	Run: func(cmd *cobra.Command, args []string) {
		var eg errgroup.Group

		servers, err := extractServerNames(args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
			return
		}
		for _, name := range servers {
			name := name
			continue
			eg.Go(func() error {
				var reqID string

				isServer, where, err := groupOrServer(name)
				//				"on":       client.PowerOnServer,
				if err != nil {
					fmt.Fprintf(os.Stderr, "ERROR (%s): %s\n", name, err)
				} else if isServer {
					if reqID, err = client.DeleteServer(where); err != nil {
						fmt.Fprintf(os.Stderr, "ERROR deleting server %s: %s\n", name, err)
					} else {
						log.Printf("Deleting server %s: %s", name, reqID)
					}
				} else {
					if reqID, err = client.DeleteGroup(where); err != nil {
						fmt.Fprintf(os.Stderr, "ERROR deleting group %s: %s\n", name, err)
					} else {
						log.Printf("Deleting group %s: %s\n", name, reqID)
					}
				}

				if reqID != "" {
					client.PollStatusFn(reqID, intvl, func(s clcv2.QueueStatus) {
						log.Printf("Deleting %s: %s", name, s)
					})
				}
				return err
			})
		}
		_ = eg.Wait()
	},
}

func init() {
	Root.AddCommand(On)
}
