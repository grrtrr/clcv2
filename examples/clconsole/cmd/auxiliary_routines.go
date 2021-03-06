package cmd

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/clcv2/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

/*
 * Helper Functions
 */

// die is like die in Perl
func die(format string, a ...interface{}) {
	format = fmt.Sprintf("%s: %s\n", path.Base(os.Args[0]), strings.TrimSpace(format))
	fmt.Fprintf(os.Stderr, format, a...)
	os.Exit(1)
}

// checkArgs returns a cobra-compatible PreRunE argument-validation function
func checkArgs(nargs int, errMsg string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) != nargs {
			return errors.Errorf(errMsg)
		}
		return nil
	}
}

// checkAtLeastArgs is analogous to checkArgs
func checkAtLeastArgs(nargs int, errMsg string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) < nargs {
			return errors.Errorf(errMsg)
		}
		return nil
	}
}

// truncate ensures that the length of @s does not exceed @maxlen
func truncate(s string, maxlen int) string {
	if len(s) >= maxlen {
		s = s[:maxlen]
	}
	return s
}

// CLCItem is a representation of a name in CLC
type CLCItem struct {
	ID       string // server name or hardware group UUID
	isServer bool   // whether @ID refers to a server
}

// groupOrServer decides whether @name refers to a CLCv2 hardware group or a server.
// It indicates the result via a returned boolean flag, and resolves @name into @id.
func groupOrServer(name string) (isServer bool, id string, err error) {
	// Strip trailing slashes that hint at a group name (but are not part of the CLC name).
	if where := strings.TrimRight(name, "/"); where == "" {
		// An emtpy name by default refers to all entries in the default data centre.
		return false, "", nil
	} else if _, errHex := hex.DecodeString(where); errHex == nil {
		/* If it decodes as a hex value, assume it is a Hardware Group UUID */
		return false, where, nil
	} else if utils.LooksLikeServerName(where) { /* Starts with a location identifier and is not hex ... */
		return true, strings.ToUpper(where), nil
	} else if conf.Location != "" { /* Fallback: assume it is a group */
		if group, err := client.GetGroupByName(where, conf.Location); err != nil {
			return false, where, errors.Errorf("failed to resolve group name %q: %s", where, err)
		} else if group == nil {
			return false, where, errors.Errorf("no group named %q was found in %s", where, conf.Location)
		} else {
			return false, group.Id, nil
		}
		return false, "", errors.Errorf("unable to resolve group name %q in %s", where, conf.Location)
	} else if conf.Location == "" {
		return false, "", errors.Errorf("%q looks like a group name - need a location (-l argument) to resolve it", where)
	} else {
		return false, "", errors.Errorf("unable to determine whether %q is a server or a group", where)
	}
}

// resolveNames resolves @args into groups/servers in parallel
func resolveNames(args []string) (groups, servers []string, err error) {
	var eg, ctx = errgroup.WithContext(context.Background())
	var ch = make(chan CLCItem, 0)

	// Do a first pass to check if there are mixed server/group arguments.
	// Correct the location based on the server name if necessary
	for _, arg := range args {
		setLocationBasedOnServerName(arg)
	}

	for _, arg := range args {
		arg := arg
		eg.Go(func() error {
			isServer, where, err := groupOrServer(arg)
			if err != nil {
				return err
			}
			select {
			case ch <- CLCItem{ID: where, isServer: isServer}:
			case <-ctx.Done():
			}
			return ctx.Err()
		})
	}

	// close channel when done
	go func() {
		eg.Wait()
		close(ch)
	}()

	for item := range ch {
		if item.isServer {
			servers = append(servers, item.ID)
		} else {
			groups = append(groups, item.ID)
		}
	}
	// Return the accumulated error
	if err = eg.Wait(); err != nil {
		return nil, nil, err
	}
	return groups, servers, nil
}

// setLocationBasedOnServerName corrects the global location value based on @serverName
func setLocationBasedOnServerName(serverName string) {
	if !utils.LooksLikeServerName(serverName) {
		/* skip */
	} else if srvLoc := utils.ExtractLocationFromServerName(serverName); conf.Location == "" {
		conf.Location = srvLoc
	} else if strings.ToUpper(conf.Location) != srvLoc {
		fmt.Fprintf(os.Stderr, "Correcting location from %q to %q for server %s\n", conf.Location, srvLoc, serverName)
		conf.Location = srvLoc
	}
}

// resolveNet attempts to resolve @s into a Network in @location.
// It supports hex ID, CIDR, IP address, or network name.
// NOTE: requires global @client to be initialized
func resolveNet(s, location string) (netw *clcv2.Network, err error) {
	if _, err = hex.DecodeString(s); err == nil {
		/* already looks like a HEX ID */
		return nil, nil
	} else if location == "" {
		return nil, errors.Errorf("need a location argument (-l) to resolve network %q", s)
	} else if _, network, err := net.ParseCIDR(s); err == nil { // CIDR string
		log.Printf("Looking up network for CIDR %s in %s", network, location)
		if netw, err = client.GetNetworkIdByCIDR(network.String(), location); err != nil {
			return nil, errors.Errorf("failed to look up CIDR %q in %s: %s",
				network, location, err)
		} else if netw == nil {
			return nil, errors.Errorf("no network matching %s found in %s",
				s, location)
		}
	} else if ip := net.ParseIP(s); ip != nil { // IP address (without CIDR netmask)
		log.Printf("Looking up network for IP %s in %s", s, location)
		if netw, err = client.GetNetworkIdByIP(s, location); err != nil {
			return nil, errors.Errorf("failed to look up IP %q in %s: %s",
				ip, location, err)
		} else if netw == nil {
			return nil, errors.Errorf("no network containing %s found in %s",
				s, location)
		}
	} else { // network name
		log.Printf("Looking up network named %q in %s", s, location)
		if netw, err = client.GetNetworkIdByName(s, location); err != nil {
			return nil, errors.Errorf("failed to look up network %q in %s: %s",
				s, location, err)
		} else if netw == nil {
			return nil, errors.Errorf("no network named %q found in %s",
				s, location)
		}
	}
	log.Printf("Found network %s with gateway %s (ID %s)", netw.Cidr, netw.Gateway, netw.Id)
	return netw, nil
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
		} else if conf.Location == "" {
			return nil, errors.Errorf("Location argument (-l) is required in order to traverse group %s", name)
		} else if root, err = client.GetGroups(conf.Location); err != nil {
			return nil, errors.Errorf("Failed to look up groups at %s: %s", conf.Location, err)
		} else {
			start := root
			if where != "" {
				start = clcv2.FindGroupNode(root, func(g *clcv2.Group) bool { return g.Id == where })
				if start == nil {
					return nil, errors.Errorf("Failed to look up group %q in %s - is the location correct?", where, conf.Location)
				}
			}
			groupVisitor(start)
		}
	}
	return ret, nil
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
