package clcv2

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/user"
	"path"
	"runtime"
	"strings"

	"github.com/pkg/errors"
)

const (
	// Name of the file to store the last bearer-token credentials
	credentialsName = "credentials.json"
)

// CLIClient specializes Client for command-line use
type CLIClient struct {
	*Client
	ClientConfig
}

// ClientConfig encapsulates a commandline-client configuration file
type ClientConfig struct {
	Username string
	Password string

	LastAccount  string // track account that was used last
	LastLocation string // track data centre that was used last
	*LoginRes
}

// NewCLIClient returns an authenticated commandline client.
// This will use the default values for AccountAlias  and LocationAlias.
// It will respect the following environment variables to override the defaults:
// - CLC_ACCOUNT:  takes precedence over default AccountAlias
// - CLC_LOCATION: takes precedence over default LocationAlias
// - CLC_BASE_URL: overrides the API URL (for testing)
func NewCLIClient(user, pass, account string) (*CLIClient, error) {
	var client = &CLIClient{Client: newClient()}

	client.LoginReq = LoginReq{user, pass}

	client.credentialsChanged = client.saveCredentials
	if Debug {
		client.Log = log.New(os.Stdout, "", log.Ltime|log.Lshortfile)
	}

	if loginRes, err := client.loadCredentials(); err != nil {
		return nil, err
	} else if loginRes != nil {
		client.credentials = loginRes
	} else if err = client.login(); err != nil {
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

	return client, nil
}

// Populate and allocate c.credentials, either by loading from file or via a fresh login.
// Save (updated) credentials if successful.
func (c *CLIClient) loadCredentials() (*LoginRes, error) {
	var credsFile = path.Join(GetClcHome(), credentialsName)

	if _, err := os.Stat(credsFile); err == nil {
		var loginRes = new(LoginRes)

		fd, err := os.Open(credsFile)
		if err != nil {
			return nil, errors.Errorf("failed to load credentials: %s", err)
		}
		defer fd.Close()

		if err = json.NewDecoder(fd).Decode(loginRes); err != nil {
			return nil, errors.Errorf("failed to deserialize %s: %s", credsFile, err)
		}

		if strings.ToLower(loginRes.User) == strings.ToLower(c.LoginReq.Username) {
			return loginRes, nil
		}
		/* User switch: move the original credentials file to a backup extension. */
		os.Rename(credsFile, credsFile+".bak")
	} else if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return nil, nil
}

// Save credentials to CLC_HOME/$credentialsName. Return error on failure.
func (c *CLIClient) saveCredentials() error {
	if c.credentials == nil {
		return errors.Errorf("login credentials not initialized")
	} else if enc, err := json.MarshalIndent(c.credentials, "", "\t"); err != nil {
		return errors.Errorf("failed to serialize bearer credentials: %s", err)
	} else {
		var clcHome = GetClcHome()

		if _, err := os.Stat(clcHome); os.IsNotExist(err) {
			if err = os.MkdirAll(clcHome, 0700); err != nil {
				return errors.Errorf("failed to create CLC home %s: %s", clcHome, err)
			}
		}
		return ioutil.WriteFile(path.Join(clcHome, credentialsName), append(enc, '\n'), 0600)
	}
}

// GetClcHome returns the path to the CLC configuration directory, which is the same
// as used by, and compatible with, clc-go-cli (including the CLC_HOME environment variable).
func GetClcHome() string {
	if clcHome := os.Getenv("CLC_HOME"); clcHome != "" {
		return clcHome
	}

	u, err := user.Current()
	if err != nil {
		log.Fatalf("failed to look up current user: %s", err)
	}

	if runtime.GOOS == "windows" {
		return path.Join(u.HomeDir, "clc")
	} else {
		return path.Join(u.HomeDir, ".clc")
	}
}
