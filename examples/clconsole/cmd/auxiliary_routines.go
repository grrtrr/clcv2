package cmd

import (
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"strings"

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

// truncate ensures that the length of @s does not exceed @maxlen
func truncate(s string, maxlen int) string {
	if len(s) >= maxlen {
		s = s[:maxlen]
	}
	return s
}

// groupOrServer decides whether @name refers to a CLCv2 hardware group or a server.
// It indicates the result via a returned boolean flag, and resolves @name into @id.
func groupOrServer(name string) (isServer bool, id string, err error) {
	// Strip trailing slashes that hint at a group name (but are not part of the CLC name).
	if where := strings.TrimRight(name, "/"); where == "" {
		// An emtpy name by default refers to all entries in the default data centre.
		return false, "", nil
	} else if _, errHex := hex.DecodeString(where); errHex == nil {
		/* If the first argument decodes as a hex value, assume it is a Hardware Group UUID */
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

// setLocationBasedOnServerName corrects the global location value based on @serverName
func setLocationBasedOnServerName(serverName string) {
	if srvLoc := utils.ExtractLocationFromServerName(serverName); conf.Location == "" {
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
