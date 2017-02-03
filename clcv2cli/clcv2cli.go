package clcv2cli

/*
 * Methods and data pertaining to commandline clients.
 */
import (
	"flag"
	"time"

	"github.com/grrtrr/clcv2"
)

// Global (commandline flag) variables
var (
	g_user, g_pass string              /* Command-line username/password */
	g_acct         string              /* Account Alias to use instead of the default */
	g_debug        bool                /* Command-line debug flag */
	g_timeout      = 180 * time.Second /* Client default timeout */
)

func init() {
	flag.StringVar(&g_user, "username", "", "CLC Login Username")
	flag.StringVar(&g_pass, "password", "", "CLC Login Password")
	flag.BoolVar(&g_debug, "d", false, "Produce debug output")
	flag.StringVar(&g_acct, "a", "", "CLC Account Alias to use (instead of default)")
	/*
	 * Caveat: keep the timeout value high, at least a few minutes.
	 *         Some operations, such as querying details of a new server immediately
	 *         after launching a CreateServer request, can take up to circa a minute.
	 */
	flag.DurationVar(&g_timeout, "timeout", 180*time.Second, "Client default timeout")
}

// NewCLIClient is a convenience wrapper around clcv2.NewCLIClient
func NewCLIClient() (*clcv2.CLIClient, error) {
	clcv2.Debug = g_debug
	clcv2.ClientTimeout = g_timeout

	return clcv2.NewCLIClient(&clcv2.ClientConfig{
		Username: g_user,
		Password: g_pass,
		Account:  g_acct,
		Location: "",
	})
}
