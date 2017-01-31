package cmd

import (
	"encoding/hex"
	"fmt"
	"os"
	"path"
	"strings"

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
		setLocationBasedOnServerName(where)
		return true, strings.ToUpper(where), nil
	} else if location != "" { /* Fallback: assume it is a group */
		if group, err := client.GetGroupByName(where, location); err != nil {
			return false, where, errors.Errorf("failed to resolve group name %q: %s", where, err)
		} else if group == nil {
			return false, where, errors.Errorf("no group named %q was found in %s", where, location)
		} else {
			return false, group.Id, nil
		}
		return false, "", errors.Errorf("unable to resolve group name %q in %s", where, location)
	} else if location == "" {
		return false, "", errors.Errorf("%q looks like a group name - need a location (-l argument) to resolve it", where)
	} else {
		return false, "", errors.Errorf("unable to determine whether %q is a server or a group", where)
	}
}

// setLocationBasedOnServerName corrects the global location value based on @serverName
func setLocationBasedOnServerName(serverName string) {
	if srvLoc := utils.ExtractLocationFromServerName(serverName); location == "" {
		fmt.Fprintf(os.Stderr, "Setting location based on server name %q\n", serverName)
		location = srvLoc
	} else if strings.ToUpper(location) != srvLoc {
		fmt.Fprintf(os.Stderr, "Correcting location from %q to %q based on server %s\n", location, srvLoc, serverName)
		location = srvLoc
	}

}
