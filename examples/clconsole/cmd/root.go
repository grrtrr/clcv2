package cmd

import (
	"os"
	"path"
	"time"

	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
	"github.com/spf13/cobra"
)

var (
	// Top-level command
	Root = &cobra.Command{Use: path.Base(os.Args[0])}

	// Client, authenticated via OnInitialize
	client *clcv2.CLIClient

	// Flags:
	conf    clcv2.ClientConfig
	debug   bool          // enable debug mode
	intvl   time.Duration // poll interval for statistics updates
	timeout time.Duration // client timeout
)

// Exit handler: ensure that the updated configuration is saved on program termination
func ExitHandler() {
	client.SaveConfig()
}

func init() {
	Root.PersistentFlags().StringVarP(&conf.Username, "username", "u", os.Getenv("CLC_USERNAME"), "CLC Login Username")
	Root.PersistentFlags().StringVarP(&conf.Password, "password", "p", os.Getenv("CLC_PASSWORD"), "CLC Login Password")
	Root.PersistentFlags().StringVarP(&conf.Account, "account", "a", os.Getenv("CLC_ACCOUNT"), "CLC account to use (instead of default)")
	Root.PersistentFlags().StringVarP(&conf.Location, "location", "l", os.Getenv("CLC_LOCATION"), "CLC data centre to use (instead of default)")

	Root.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Produce debug output")
	Root.PersistentFlags().DurationVarP(&intvl, "poll-interval", "i", 1*time.Second, "Poll interval for status updates (use 0 to disable)")
	Root.PersistentFlags().DurationVar(&timeout, "timeout", 180*time.Second, "Client default timeout")

	// Initialize client needed by the sub-commands
	cobra.OnInitialize(func() {
		var err error

		clcv2.Debug = debug
		clcv2.ClientTimeout = timeout

		client, err = clcv2.NewCLIClient(&conf)
		if err != nil {
			exit.Errorf("failed to initialize client: %s", err)
		}
		// Set the fallback data centre if no location was given
		if conf.Location == "" {
			conf.Location = client.LocationAlias
		}
	})
}
