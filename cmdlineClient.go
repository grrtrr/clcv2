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

	if loginRes, err := loadCredentials(user); err != nil {
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
func loadCredentials(expectedUser string) (*LoginRes, error) {
	var path = defaultCredentialsPath()

	if _, err := os.Stat(path); err == nil {
		var loginRes = new(LoginRes)
		fd, err := os.Open(path)
		if err != nil {
			return nil, errors.Errorf("failed to load credentials: %s", err)
		}
		defer fd.Close()

		if err = json.NewDecoder(fd).Decode(loginRes); err != nil {
			return nil, errors.Errorf("failed to deserialize %s: %s", path, err)
		}

		if strings.ToLower(loginRes.User) == strings.ToLower(expectedUser) {
			/* Found credentials and user has not changed. No login required. */
			return loginRes, nil
		}
		/* User switch: move the original credentials file to a backup extension. */
		os.Rename(path, path+".bak")
	} else if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return nil, nil
}

// Save credentials to default file path. Return error on failure.
func (c *CLIClient) saveCredentials() error {
	var credsFilePath = defaultCredentialsPath()

	if c.credentials == nil {
		return errors.Errorf("login credentials not initialized")
	} else if enc, err := json.MarshalIndent(c.credentials, "", "\t"); err != nil {
		return errors.Errorf("failed to serialize bearer credentials: %s", err)
	} else {
		var credsDir = path.Dir(credsFilePath)

		if _, err := os.Stat(credsFilePath); os.IsNotExist(err) {
			if err = os.MkdirAll(credsDir, 0700); err != nil {
				return errors.Errorf("failed to generate CLC home location: %s", err)
			}
		}
		return ioutil.WriteFile(credsFilePath, append(enc, '\n'), 0600)
	}
}

// Return the default path for commandline-client credentials file.
// It defaults to $HOME/.clc/credentials.json.
// The directory can be overriden via the $CLC_HOME environment variable.1
func defaultCredentialsPath() string {
	var clcHome = os.Getenv("CLC_HOME")

	if clcHome == "" {
		u, err := user.Current()
		if err != nil {
			log.Fatalf("failed to look up current user: %s", err)
		}
		clcHome = path.Join(u.HomeDir, ".clc")

		// Backwards-compatibility
		oldLocation := path.Join(u.HomeDir, ".clc_credentials.json")
		if _, err := os.Stat(oldLocation); err == nil {
			return oldLocation
		}
	}
	return path.Join(clcHome, "credentials.json")
}
