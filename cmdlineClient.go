package clcv2

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/user"
	"path"
	"strings"

	"github.com/pkg/errors"
)

// CLIClient specializes Client for command-line use
type CLIClient struct {
	*Client
}

// NewCLIClient returns an authenticated commandline client.
// This will use the default values for AccountAlias  and LocationAlias.
// It will respect the following environment variables to override the defaults:
// - CLC_ACCOUNT:  takes precedence over default AccountAlias
// - CLC_LOCATION: takes precedence over default LocationAlias
// - CLC_BASE_URL: overrides the API URL (for testing)
func NewCLIClient(user, pass, account string) (*CLIClient, error) {
	var client = &CLIClient{initClient(user, pass)}

	client.credentialsChanged = client.saveCredentials
	if Debug {
		client.Log = log.New(os.Stdout, "", log.Ltime|log.Lshortfile)
	}

	// Set/override %baseURL (experimental).
	if envURL := os.Getenv("CLC_BASE_URL"); envURL != "" {
		url, err := url.Parse(envURL)
		if err != nil {
			return nil, err
		}
		if url.Scheme == "" {
			url.Scheme = "https"
		}
		baseURL = url.String()
	}

	if err := client.loadCredentials(); err != nil {
		return nil, err
	}

	// Set/override AccountAlias
	if account != "" {
		client.AccountAlias = account
	} else if account = os.Getenv("CLC_ACCOUNT"); account != "" {
		client.AccountAlias = account
	} else { // may have been initialized from disk
		client.AccountAlias = client.credentials.AccountAlias
	}

	// Set/override LocationAlias
	if location := os.Getenv("CLC_LOCATION"); location != "" {
		client.LocationAlias = location
	} else {
		client.LocationAlias = client.credentials.LocationAlias
	}

	return client, nil
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
		return errors.Errorf("login credentials not initialized")
	} else if enc, err := json.MarshalIndent(c.credentials, "", "\t"); err != nil {
		return err
	} else {
		return ioutil.WriteFile(defaultCredentialsPath(), append(enc, '\n'), 0600)
	}
}

// Return the default path for commandline-client credentials file.
func defaultCredentialsPath() string {
	if env := os.Getenv("CLC_CREDENTIALS"); env != "" {
		return env
	}
	u, err := user.Current()
	if err != nil {
		panic(errors.Errorf("failed to look up current user: %s", err))
	}
	return path.Join(u.HomeDir, ".clc_credentials.json")
}

// Remove (stale) credentials
func (c *CLIClient) destroyCredentials() {
	os.Remove(defaultCredentialsPath())
}
