package clcv2

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"net/http/httputil"
	"reflect"
	"strings"
	"time"

	"github.com/PuerkitoBio/rehttp"
	"github.com/Sirupsen/logrus"
)

const (
	// CenturyLink Cloud main v2 API url
	BaseURL = "https://api.ctl.io"

	// Maximum number of retries per request.
	MaxRetries = 3

	// Per-request retry delay for the retryer.
	StepDelay = time.Second * 10
)

// GLOBAL VARIABLES
var (
	// allow overriding of the %BaseURL default
	baseURL = BaseURL
)

// Client wraps a http.Client, along with credentials and logging information.
type Client struct {
	// Login credentials
	LoginReq

	// AccountAlias to use (defaults to @credentials.AccountAlias, but can be overridden)
	AccountAlias string

	// Logger used for (debugging) output.
	Log logrus.StdLogger

	/*
	 * private
	 */
	// Authentication information (may be updated when the BearerToken is stale)
	credentials *LoginRes

	// Performs the actual requests
	requestor *http.Client

	// Optional context to control cancellation/deadline
	ctx context.Context

	// controls automatic re-login
	retryingLogin bool

	// Optional callback which is called when @credentials is updated
	credentialsChanged func() error
}

// LoginReq is the data structure required to perform the initial CLCv2 login request.
type LoginReq struct {
	// Control Portal user name.
	Username string `json:"username"`

	// Control Portal password.
	Password string `json:"password"`
}

// LoginRes is the data returned by the CLCv2 endpoint after successful authentication.
type LoginRes struct {
	// Control Portal user name value
	User string `json:"userName"`

	// Account that contains this user record
	AccountAlias string `json:"accountAlias"`

	// Default data center of the user
	LocationAlias string `json:"locationAlias"`

	// Permission roles associated with this user
	Roles []string `json:"roles"`

	// Security token for this user that is included in the Authorization header
	// for all other API requests as "Bearer [LONG TOKEN VALUE]".
	BearerToken string `json:"bearerToken"`
}

func (l LoginRes) String() string {
	return fmt.Sprintf("User=%s, Account=%s, Location=%s, Roles=%s", l.User,
		l.AccountAlias, l.LocationAlias, strings.Join(l.Roles, ", "))
}

// NewClient returns an initialized client, performing the login request.
func NewClient(user, pass string) (*Client, error) {
	client := initClient(user, pass)
	if err := client.login(); err != nil {
		return nil, err
	}
	return client, nil
}

// SetContext sets the cancellation context of @c to @ctx
func (c *Client) SetContext(ctx context.Context) {
	c.ctx = ctx
}

// initClient initializes the parts common to both Client and CLIClient
func initClient(user, pass string) *Client {
	client := &Client{LoginReq: LoginReq{user, pass}}
	client.requestor = &http.Client{
		Transport: rehttp.NewTransport(nil, // default transport
			client.retryer(MaxRetries),
			// Note: using g_timeout as upper bound for the exponential backoff.
			//       This means g_timeout has to be large enough to run MaxRetries
			//       requests with individual retries.
			rehttp.ExpJitterDelay(StepDelay, g_timeout),
		),
	}
	return client
}

// Log in and update credentials if successful.
func (c *Client) login() error {
	if c.LoginReq.Username == "" || c.LoginReq.Password == "" {
		return fmt.Errorf("invalid CLC credentials %q/%q", c.LoginReq.Username, c.LoginReq.Password)
	}
	c.credentials = new(LoginRes)
	if err := c.getCLCResponse("POST", "/v2/authentication/login", &c.LoginReq, c.credentials); err != nil {
		return err
	}
	c.AccountAlias = c.credentials.AccountAlias

	if c.credentialsChanged != nil {
		return c.credentialsChanged()
	}
	return nil
}

// retryer implements the retry policy: (a) any failure, (b) temporary failure status codes
func (c *Client) retryer(maxRetries int) rehttp.RetryFn {
	return rehttp.RetryFn(func(at rehttp.Attempt) bool {
		if c.ctx != nil && c.ctx.Err() != nil {
			return false
		}
		if at.Index < maxRetries {
			if at.Response == nil {
				if c.Log != nil {
					c.Log.Printf("request failed - retry #%d", at.Index+1)
				}
				return true
			}
			/* Request timeout, server error, bad gateway, service unavailable, gateway timeout */
			switch at.Response.StatusCode {
			case 408, 500, 502, 503, 504:
				if c.Log != nil {
					c.Log.Printf("request returned %q - retry #%d", at.Response.Status, at.Index+1)
				}
				return true
			}
		}
		return false
	})
}

// getCLCResponse performs a CLC v2 main API request
// @verb: Http verb to use
// @path: relative to BaseURL (includes the 'v2' version).
func (c *Client) getCLCResponse(verb, path string, reqModel, resModel interface{}) (err error) {
	return c.getResponse(baseURL+path, verb, reqModel, resModel)
}

// getResponse performs a generic request
// @url:  request URL
// @verb: request verb// @reqModel: request model to serialize, or nil.
// @resModel: result model to deserialize, must be a pointer to the expected result, or nil.
// Evaluates the StatusCode of the BaseResponse (embedded) in @inModel and sets @err accordingly.
// If @err == nil, fills in @resModel, else returns error.
func (c *Client) getResponse(url, verb string, reqModel, resModel interface{}) (err error) {
	var reqBody io.Reader

	if reqModel != nil {
		if g_debug && c.Log != nil {
			c.Log.Printf("reqModel %T %+v\n", reqModel, reqModel)
		}

		jsonReq, err := json.Marshal(reqModel)
		if err != nil {
			return fmt.Errorf("failed to encode request model %T %+v: %s", reqModel, reqModel, err)
		}
		reqBody = bytes.NewBuffer(jsonReq)
	}

	/* resModel must be a pointer type (call-by-value) */
	if resModel != nil {
		if resType := reflect.TypeOf(resModel); resType.Kind() != reflect.Ptr {
			return fmt.Errorf("Expecting pointer to result model %T", resModel)
		} else if g_debug && c.Log != nil {
			c.Log.Printf("resModel %T %+v", resModel, resModel)
		}
	}

	req, err := http.NewRequest(verb, url, reqBody)
	if err != nil {
		return
	} else if c.ctx != nil {
		req = req.WithContext(c.ctx)
	}

	if c.credentials != nil && c.credentials.BearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.credentials.BearerToken)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json")

	if g_debug && c.Log != nil {
		reqDump, _ := httputil.DumpRequest(req, true)
		c.Log.Printf("%s", reqDump)
	}

	res, err := c.requestor.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if g_debug && c.Log != nil {
		resDump, _ := httputil.DumpResponse(res, true)
		c.Log.Printf("%s", resDump)
	}

	switch res.StatusCode {
	case 200, 201, 202, 204: /* OK / CREATED / ACCEPTED / NO CONTENT */
		if resModel != nil {
			if res.ContentLength == 0 {
				return fmt.Errorf("Unable do populate %T result model, due to empty %q response",
					resModel, res.Status)
			}
			return json.NewDecoder(res.Body).Decode(resModel)
		} else if res.ContentLength > 0 {
			return fmt.Errorf("Unable to decode non-empty %q response (%d bytes) to nil response model",
				res.Status, res.ContentLength)
		}
		return nil
	case 401:
		// This is returned if the BearerToken is missing or has become stale.
		if c.retryingLogin {
			return errors.New("failed to re-authenticate, credentials may be invalid")
		}
		if _, isLoginReq := reqModel.(*LoginReq); !isLoginReq {
			if g_debug && c.Log != nil {
				log.Printf("credentials are stale, retrying login ...")
			}
			// FIXME: the following is not thread-safe (multiple concurrent clients):
			c.retryingLogin = true
			if err = c.login(); err != nil {
				return err
			}
			if g_debug && c.Log != nil {
				log.Printf("re-authentication worked, retrying request ...")
			}
			if err = c.getResponse(url, verb, reqModel, resModel); err != nil {
				return err
			}
			c.retryingLogin = false
			return nil
		}
		return errors.New("authentication credentials are stale or invalid.")
	}

	// Remaining error cases: res.ContentLength is not reliable - in the SBS case, it used
	// Transfer-Encoding "chunked", without a Content-Length.
	body, err := ioutil.ReadAll(res.Body)
	if err != nil && res.ContentLength > 0 {
		return fmt.Errorf("failed to read error response %d body: %s", res.StatusCode, err)
	} else if len(body) > 0 {
		//
		// Currently 5 different types of response have been observed:
		// 1) bare JSON string
		// 2) struct { message: "string" }
		// 3) struct { message: "string", "modelState": map[string]interface{} }
		//    E.g.:  {"":["The server must be in Active or Archived state."]}
		//	      "modelState":{"body.networkId":["The network vlan_1249_10.81.149 is not valid."]}
		//	      "modelState":{"":["The server must be in Active or Archived state."]}
		// 4) struct { error: "string" }, e.g. { "error":"Missing required parameter: serverId"}
		// 5) struct { error: "string", validationMessages: ["string"] } - like (4), with array of messages
		//
		errMsg := string(body)
		if ct, _, _ := mime.ParseMediaType(res.Header.Get("Content-Type")); ct == "application/json" {
			/* Code thanks to & inspired by clc-go-cli */
			var payload map[string]interface{}
			var sbsError struct {
				Error              string
				ValidationMessages []string
			}

			if err := json.Unmarshal(body, &payload); err != nil {
				/* Failed to decode as struct, try string (1) next. */
				if err = json.Unmarshal(body, &errMsg); err != nil {
					errMsg = string(body)
				}
			} else if errors, ok := payload["modelState"]; ok {
				if bytes, err := json.Marshal(errors); err == nil {
					errMsg = string(bytes)
				}
			} else if errors, ok := payload["message"]; ok {
				if msg, ok := errors.(string); ok {
					errMsg = msg
				}
			} else if err = json.Unmarshal(body, &sbsError); err == nil {
				errMsg = fmt.Sprintf("Error - %s", sbsError.Error)
				if len(sbsError.ValidationMessages) > 0 {
					errMsg += fmt.Sprintf(" Details: %q", strings.Join(sbsError.ValidationMessages, ", "))
				}
			} else if error, ok := payload["error"]; ok {
				if msg, ok := error.(string); ok {
					errMsg = fmt.Sprintf("Error - %s", msg)
				}
			}
		}
		err = fmt.Errorf("%s (status: %d)", errMsg, res.StatusCode)
	} else {
		err = errors.New(res.Status)
	}
	return
}
