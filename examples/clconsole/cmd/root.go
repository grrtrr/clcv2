package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/clcv2/clcv2cli"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	Root = &cobra.Command{
		Use: "clconsole",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
			// Initialize client needed by the sub-commands
			clcv2.Debug = debug
			clcv2.ClientTimeout = timeout

			username, password, err := clcv2cli.ResolveUserAndPass(user, pass)
			if err != nil {
				return errors.Errorf("failed to resolve username/password: %s", err)
			}

			client, err = clcv2.NewCLIClient(username, password, account)
			if err != nil {
				return errors.Errorf("failed to initialize client: %s", err)
			}

			if location == "" {
				var action = cmd.Name()
				if action == "networks" || action == "show" || action == "templates" {
					fmt.Fprintf(os.Stderr, "Note: no location argument (-l) given, defaulting to %s.\n", client.LocationAlias)
				}
				location = client.LocationAlias
			}

			return nil
		},
	}

	// Client, authenticated via PersistentPreRunE
	client *clcv2.CLIClient

	// Flags:

	user, pass string // username and password
	account    string // default account alias
	location   string // default data centre location

	debug bool // enable debug mode

	intvl   time.Duration // poll interval for statistics updates
	timeout time.Duration // client timeout

)

func init() {
	Root.PersistentFlags().StringVarP(&user, "username", "u", os.Getenv("CLC_USERNAME"), "CLC Login Username")
	Root.PersistentFlags().StringVarP(&pass, "password", "p", os.Getenv("CLC_PASSWORD"), "CLC Login Password")
	Root.PersistentFlags().StringVarP(&account, "account", "a", os.Getenv("CLC_ACCOUNT"), "CLC Login Password")
	Root.PersistentFlags().StringVarP(&location, "location", "l", os.Getenv("CLC_LOCATION"), "CLC Account Alias to use (instead of default)")

	Root.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Produce debug output")
	Root.PersistentFlags().DurationVarP(&intvl, "poll-interval", "i", 1*time.Second, "Poll interval for status updates (use 0 to disable)")
	Root.PersistentFlags().DurationVar(&timeout, "timeout", 180*time.Second, "Client default timeout")
}
