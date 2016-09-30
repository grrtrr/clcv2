package clcv2

/*
 * Methods and data pertaining to commandline clients.
 */
import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/user"
	"path"
	"strings"
	"time"

	"github.com/grrtrr/clcv2/utils"
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

// CLIClient specializes Client for command-line use
type CLIClient struct {
	*Client
}

// NewCLIClient returns an authenticated commandline client.
// This will use the default values for AccountAlias  and LocationAlias.
// It will respect the following environment variable to override the defaults:
// - CLC_ACCOUNT: takes precedence over default AccountAlias
func NewCLIClient() (client *CLIClient, err error) {
	username, password, err := resolveUserAndPass()
	if err != nil {
		return nil, err
	}
	client = &CLIClient{initClient(username, password)}

	client.credentialsChanged = client.saveCredentials

	if g_debug {
		client.Log = log.New(os.Stdout, "", log.Ltime|log.Lshortfile)
	}

	if err = setBaseURL(); err != nil {
		return nil, err
	}

	if err = client.loadCredentials(); err != nil {
		return nil, err
	}

	/* Set/override account alias. */
	client.AccountAlias = client.credentials.AccountAlias

	if account := os.Getenv("CLC_ACCOUNT"); account != "" {
		client.AccountAlias = account
	}

	/* Commandline flags take precedence over environment variables. */
	if g_acct != "" {
		client.AccountAlias = g_acct
	}
	return client, nil
}

// setBaseURL sets the URL base based on $CLC_BASE_URL.
func setBaseURL() error {
	if envURL := os.Getenv("CLC_BASE_URL"); envURL != "" {
		url, err := url.Parse(envURL)
		if err != nil {
			return err
		}
		if url.Scheme == "" {
			url.Scheme = "https"
		}
		baseURL = url.String()
	}
	return nil
}

// Populate and allocate c.credentials, either by loading from file or via a fresh login.
// Save (updated) credentials if successful.
func (c *CLIClient) loadCredentials() error {
	var path = defaultCredentialsPath()

	if _, err := os.Stat(path); err == nil {
		fd, err := os.Open(path)
		if err != nil {
			return err
		}
		defer fd.Close()

		c.credentials = new(LoginRes)
		if err = json.NewDecoder(fd).Decode(c.credentials); err != nil {
			return err
		}

		if strings.ToLower(c.credentials.User) == strings.ToLower(c.LoginReq.Username) {
			/* Found credentials and user has not changed. No login required. */
			return nil
		}
		/* User switch: move the original credentials file to a backup extension. */
		os.Rename(path, path+".bak")
	} else if err != nil && !os.IsNotExist(err) {
		return err
	}

	return c.login()
}

// Save credentials to default file path. Return error on failure.
func (c *CLIClient) saveCredentials() error {
	if c.credentials == nil {
		return fmt.Errorf("login credentials not initialized")
	}
	enc, err := json.MarshalIndent(c.credentials, "", "\t")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(defaultCredentialsPath(), append(enc, '\n'), 0600)
}

// Remove (stale) credentials
func (c *CLIClient) destroyCredentials() {
	os.Remove(defaultCredentialsPath())
}

/*
 * Auxiliary Functions
 */

// Return the default path for commandline-client credentials file.
func defaultCredentialsPath() string {
	if env := os.Getenv("CLC_CREDENTIALS"); env != "" {
		return env
	}
	u, err := user.Current()
	if err != nil {
		panic(fmt.Errorf("failed to look up current user: %s", err))
	}
	return path.Join(u.HomeDir, ".clc_credentials.json")
}

/**
 * Support multiple ways of resolving the username and password
 * 1. directly (pass-through),
 * 2. command-line flags (g_user, g_pass),
 * 3. environment variables (CLC_USERNAME, CLC_PASSWORD),
 * 4. prompt for values
 */
func resolveUserAndPass() (username, password string, err error) {
	var prompt string = "Username"

	username = g_user
	if username == "" {
		username = os.Getenv("CLC_USERNAME")
	}
	if username == "" {
		if username, err = utils.PromptInput(prompt); err != nil {
			return
		}
		prompt = "Password"
	} else {
		prompt = fmt.Sprintf("Password for %s", username)
	}

	password = g_pass
	if password == "" {
		password = os.Getenv("CLC_PASSWORD")
	}
	if password == "" {
		if password, err = utils.GetPass(prompt); err != nil {
			return
		}
	}
	return
}
