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

	yaml "gopkg.in/yaml.v2"

	"github.com/grrtrr/clcv2/utils"
	"github.com/pkg/errors"
)

const (
	// Name of the file to store the last bearer-token credentials
	credentialsName = "credentials.json"

	// Configuration file in CLC_HOME that stores the ClientConfig
	configName = "client_config.yml"
)

// CLIClient specializes Client for command-line use
type CLIClient struct {
	*Client
	Config *ClientConfig
}

// ClientConfig encapsulates a commandline-client configuration file
type ClientConfig struct {
	Username     string `yaml:"User"`        // CLC portal username
	Password     string `yaml:"Password"`    // CLC portal password (FIXME: store encrypted)
	LastAccount  string `yaml:"Account"`     // account that was used last time
	LastLocation string `yaml:"Data Centre"` // data centre that was used last time
}

// NewCLIClient returns an authenticated commandline client.
// This will use the default values for AccountAlias  and LocationAlias.
// It will respect the following environment variables to override the defaults:
// - CLC_ACCOUNT:  takes precedence over default AccountAlias
// - CLC_LOCATION: takes precedence over default LocationAlias
// - CLC_BASE_URL: overrides the API URL (for testing)
func NewCLIClient(conf *ClientConfig) (*CLIClient, error) {
	// Attempt to load existing configuration first, and reconcile with @conf.
	savedConfig, err := LoadClientConfig()
	if err != nil && conf == nil {
		return nil, errors.Errorf("failed to load saved configuration: %s", err)
	}

	if conf == nil {
		conf = savedConfig
	} else if conf.Username == "" && savedConfig != nil {
		conf.Username, conf.Password = savedConfig.Username, savedConfig.Password
	}

	if conf == nil {
		conf = &ClientConfig{}
	}

	// Ensure that both username and password were filled in
	conf.Username, conf.Password = utils.ResolveUserAndPass(conf.Username, conf.Password)

	client := &CLIClient{
		Client: newClient(conf.Username, conf.Password),
		Config: conf,
	}
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
	if conf.LastAccount != "" {
		client.AccountAlias = conf.LastAccount
	} else if account := os.Getenv("CLC_ACCOUNT"); account != "" {
		client.AccountAlias = account
	} else {
		client.AccountAlias = client.credentials.AccountAlias
	}

	// Set/override LocationAlias
	if conf.LastLocation != "" {
		client.LocationAlias = conf.LastLocation
	} else if location := os.Getenv("CLC_LOCATION"); location != "" {
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

// LoadClientConfig attempts to load a configuration from CLC_HOME/configName
func LoadClientConfig() (*ClientConfig, error) {
	var confFile = path.Join(GetClcHome(), configName)

	if _, err := os.Stat(confFile); err == nil {
		var config ClientConfig

		fd, err := os.Open(confFile)
		if err != nil {
			return nil, errors.Errorf("failed to load client config: %s", err)
		}
		defer fd.Close()

		if content, err := ioutil.ReadAll(fd); err != nil {
			return nil, errors.Errorf("failed to read %s: %s", confFile, err)
		} else if err = yaml.Unmarshal(content, &config); err != nil {
			return nil, errors.Errorf("failed to deserialize %s: %s", confFile, err)
		}
		return &config, nil
	}
	return configFromCliGo()
}

// configFromCliGo checks to see if a clc-cli-go configuration file exists.
// If yes, it will import a client configuration based on those settings.
func configFromCliGo() (*ClientConfig, error) {
	var confFile = path.Join(GetClcHome(), "config.yml")

	if _, err := os.Stat(confFile); err == nil {
		var cliGoData = make(map[string]interface{})

		fd, err := os.Open(confFile)
		if err != nil {
			return nil, errors.Errorf("failed to load client config: %s", err)
		}
		defer fd.Close()

		if content, err := ioutil.ReadAll(fd); err != nil {
			return nil, errors.Errorf("failed to read %s: %s", confFile, err)
		} else if err = yaml.Unmarshal(content, cliGoData); err != nil {
			return nil, errors.Errorf("failed to deserialize %s: %s", confFile, err)
		}

		return &ClientConfig{
			Username:     cliGoData["user"].(string),
			Password:     cliGoData["password"].(string),
			LastLocation: cliGoData["defaultdatacenter"].(string),
		}, nil
	}
	return nil, nil
}

// SaveConfig writes the configuration data of @c to CLC_HOME/configName
func (c *CLIClient) SaveConfig() error {
	if c == nil || c.Client == nil {
		return errors.New("attempt to save configuration for nil client")
	} else if c.Config == nil {
		return errors.New("attempt to save a nil client configuration")
	}

	if enc, err := yaml.Marshal(c.Config); err != nil {
		return errors.Errorf("failed to serialize client configuration: %s", err)
	} else {
		return writeCLCdata(configName, enc, 0644)
	}
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
		// FIXME: if names differ, save to credsFile + loginRes.User
		/* User switch: move the original credentials file to a backup extension. */
		os.Rename(credsFile, credsFile+".bak")
	} else if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return nil, nil
}

// Save credentials to CLC_HOME/$credentialsName. Return error on failure.
func (c *CLIClient) saveCredentials() error {
	if c.credentials == nil { // nothing to serialize
		return nil
	} else if enc, err := json.MarshalIndent(c.credentials, "", "\t"); err != nil {
		return errors.Errorf("failed to serialize bearer credentials: %s", err)
	} else {
		return writeCLCdata(credentialsName, append(enc, '\n'), 0600)
	}
}

// writeCLCitem writes @data to CLC_HOME/fileName
func writeCLCdata(fileName string, data []byte, perm os.FileMode) error {
	var clcHome = GetClcHome()

	if _, err := os.Stat(clcHome); os.IsNotExist(err) {
		if err = os.MkdirAll(clcHome, 0700); err != nil {
			return errors.Errorf("failed to create CLC directory %s: %s", clcHome, err)
		}
	}
	return ioutil.WriteFile(path.Join(clcHome, fileName), data, perm)
}
