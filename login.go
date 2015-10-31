package clcv2

import (
	"terminal"
	"time"
	"flag"
	"fmt"
	"os"
)

/* Global variables */
var g_user, g_pass	string		/* Command-line username/password */
var g_timeout		time.Duration	/* Client default timeout */
var g_debug		bool     /* Command-line debug fl5Dag */

func init() {
	flag.BoolVar(&g_debug,  "d", false, "Produce debug output")
	flag.StringVar(&g_user, "u", "",    "CLC Login Username")
	flag.StringVar(&g_pass, "p", "",    "CLC Login Password")
	flag.DurationVar(&g_timeout, "timeout", 10 * time.Second, "Client default timeout")
}

type LoginReq struct {
	// Control Portal user name value
	Username	string	`json:"username"`
	// Control Portal password value.
	Password	string	`json:"password"`
}

type LoginRes struct {
	// Control Portal user name value
	UserName      string	`json: "userName"`

	// Account that contains this user record
	AccountAlias  string	`json: "accountAlias"`

	// Default data center of the user
	LocationAlias string	`json: "locationAlias"`

	// Permission roles associated with this user
	Roles         []string	`json: "roles"`

	// Security token for this user that is included in the Authorization header
	// for all other API requests as "Bearer [LONG TOKEN VALUE]".
	BearerToken   string	`json: "bearerToken"`
}

// Perform login request, fill in c.LoginRes
func (c *Client) login() (err error) {
	username, password, err := resolveUserAndPass()
	if err != nil {
		return
	}
	c.LoginRes = new(LoginRes)
	err = c.getResponse("POST", "/v2/authentication/login", &LoginReq{ username, password }, c.LoginRes)
	return
}


/**
 * Support multiple ways of resolving the username and password
 * 1. directly (pass-through),
 * 2. command-line flags (g_user, g_pass),
 * 3. environment variables (CLC_V2_API_USERNAME/PASSWORD),
 * 4. prompt for values
 */
func resolveUserAndPass() (username, password string, err error) {
	var prompt string = "Username"

	username = g_user
	if username == "" {
		username = os.Getenv("CLC_V2_API_USERNAME")
	}
	if username == "" {
		if username, err = terminal.PromptInput(prompt); err != nil {
			return
		}
		prompt = "Password"
	} else {
		prompt = fmt.Sprintf("Password for %s", username)
	}

	password = g_pass
	if password == "" {
		password = os.Getenv("CLC_V2_API_PASSWORD")
	}
	if password == "" {
		if password, err = terminal.GetPass(prompt); err != nil {
			return
		}
	}
	return
}
