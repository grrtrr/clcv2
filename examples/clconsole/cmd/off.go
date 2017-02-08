package cmd

import "github.com/spf13/cobra"

// offFlags groups the flags pertaining to powering off, shutting down, or pausing a server
var offFlags struct {
	pause bool // whether to pause server instead of shutting it down
	hard  bool // whether to do a hard power-off instead of an OS-level shutdown
}

func init() {
	// The off-command is the only visible command; 'pause' and 'shutdown' are available, but hidden
	var powerOff = &cobra.Command{
		Use:   "off  [group|server [group|server]...]",
		Short: "Power-off or suspend server(s)",
		Long:  "Shutdown, power-off (if --hard is set), or pause (if --pause is set) a server",
		Run: func(cmd *cobra.Command, args []string) {
			if offFlags.pause {
				serverCmd("pause", client.PauseServer, args)
			} else if offFlags.hard {
				serverCmd("hard power-off", client.PowerOffServer, args)
			} else {
				serverCmd("shutdown", client.ShutdownServer, args)
			}
		},
	}

	powerOff.Flags().BoolVar(&offFlags.pause, "pause", false, "Whether to suspend (pause) server instead of shutting it down")
	powerOff.Flags().BoolVar(&offFlags.hard, "hard", false, "Whether to use a hard power-off instead of an OS-level shutdown")

	Root.AddCommand(powerOff)

	Root.AddCommand(&cobra.Command{
		Use:     "pause  [group|server [group|server]...]",
		Hidden:  true, /* already included in 'off' */
		Aliases: []string{"suspend"},
		Short:   "Pause server(s)",
		Long:    "Pause (suspend) server(s); can be resumed via 'on'",
		Run: func(cmd *cobra.Command, args []string) {
			serverCmd("pause", client.PauseServer, args)
		}})

	Root.AddCommand(&cobra.Command{
		Use:     "shutdown  [group|server [group|server]...]",
		Hidden:  true, /* already included in 'off' */
		Aliases: []string{"halt"},
		Short:   "Shutdown server(s)",
		Long:    "Soft (OS-level) shutdown, followed by power-off",
		Run: func(cmd *cobra.Command, args []string) {
			serverCmd("shutdown", client.ShutdownServer, args)
		}})

}
