package clcv2

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"strings"

	"github.com/grrtrr/clcv2/utils"
)

/*
 * User Authentication
 */
type LoginReq struct {
	// Control Portal user name value.
	Username string `json:"username"`

	// Control Portal password value.
	Password string `json:"password"`
}

type LoginRes struct {
	// Control Portal user name value
	UserName string `json: "userName"`

	// Account that contains this user record
	AccountAlias string `json: "accountAlias"`

	// Default data center of the user
	LocationAlias string `json: "locationAlias"`

	// Permission roles associated with this user
	Roles []string `json: "roles"`

	// Security token for this user that is included in the Authorization header
	// for all other API requests as "Bearer [LONG TOKEN VALUE]".
	BearerToken string `json: "bearerToken"`
}

// Log in and save credentials if successful. Requires c.LoginRes.BearerToken to be empty.
func (c *Client) login(user, pass string) error {
	err := c.getResponse("POST", "/v2/authentication/login",
		&LoginReq{user, pass}, &c.LoginRes)
	if err != nil {
		return err
	}
	return c.saveCredentials()
}

// Remove (stale) credentials
func (c *Client) destroyCredentials() {
	os.Remove(defaultCredentialsPath())
}

// Try to load credentials from file at default path, or do a fresh login.
func (c *Client) loadCredentials() error {
	var path = defaultCredentialsPath()

	username, password, err := resolveUserAndPass()
	if err != nil {
		return err
	}

	if _, err := os.Stat(path); err == nil {
		fd, err := os.Open(path)
		if err != nil {
			return err
		}
		defer fd.Close()

		err = json.NewDecoder(fd).Decode(&c.LoginRes)
		if err != nil {
			return err
		}

		if strings.ToLower(username) == strings.ToLower(c.LoginRes.UserName) {
			return nil
		}
		/* User switch - clear all related environment variables */
		os.Unsetenv("CLC_ALIAS")
		os.Unsetenv("CLC_ACCOUNT")
		os.Rename(path, path+".bak")
	} else if err != nil && !os.IsNotExist(err) {
		return err
	}
	return c.login(username, password)
}

// Save credentials to default file path. Return error on failure.
func (c *Client) saveCredentials() error {
	enc, err := json.MarshalIndent(&c.LoginRes, "", "\t")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(defaultCredentialsPath(), append(enc, '\n'), 0600)
}

// Return the default path for file credentials
func defaultCredentialsPath() string {
	if env := os.Getenv("CLC_CREDENTIALS"); env != "" {
		return env
	}
	u, err := user.Current()
	if err != nil {
		panic(fmt.Errorf("Failed to look up current user: %s", err))
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
