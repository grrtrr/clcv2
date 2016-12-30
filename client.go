package clcv2

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/http/httputil"
	"reflect"
	"strings"
	"time"

	"github.com/PuerkitoBio/rehttp"
	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
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
	// Errors returned by this package
	ErrCredentialsInValid = errors.New("authentication credentials are stale or invalid")

	// Set this to true for debugging output
	Debug bool

	// Upper bound on client operations - default timeout value
	ClientTimeout = 180 * time.Second

	// allow overriding of the %BaseURL default
	baseURL = BaseURL
)

// Client wraps a http.Client, along with credentials and logging information.
type Client struct {
	// Login credentials
	LoginReq

	// AccountAlias to use (defaults to @credentials.AccountAlias, but can be overridden)
	AccountAlias string

	// Similar to @AccountAlias, set the default data center
	LocationAlias string

	// Logger used for (debugging) output.
	Log logrus.StdLogger

	/*
	 * private
	 */
	// Authentication information (may be updated when the BearerToken is stale)
	credentials *LoginRes

	// Performs the actual requests
	requestor *http.Client

	// Cancellation context (used by @cancel). Can be overridden via SetContext()
	ctx context.Context

	// Cancels this client via @ctx
	cancel context.CancelFunc

	// controls automatic re-login
	retryingLogin bool

	// Optional callback which is called when @credentials is updated
	credentialsChanged func() error
}

// RegisteredAccountAlias is a read-only get method for credentials.AccountAlias
func (c *Client) RegisteredAccountAlias() string {
	return c.credentials.AccountAlias
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
	c.ctx, c.cancel = context.WithCancel(ctx)
}

// Cancel cancels @c.ctx
func (c *Client) Cancel() {
	c.cancel()
}

// Context returns the cancellation context of @c
func (c *Client) Context() context.Context {
	return c.ctx
}

// initClient initializes the parts common to both Client and CLIClient
func initClient(user, pass string) *Client {
	var client = &Client{LoginReq: LoginReq{user, pass}}

	client.requestor = &http.Client{
		Transport: rehttp.NewTransport(nil, // default transport
			client.retryer(MaxRetries),
			// Note: using ClientTimeout as upper bound for the exponential backoff.
			//       This means g_timeout has to be large enough to run MaxRetries
			//       requests with individual retries.
			rehttp.ExpJitterDelay(StepDelay, ClientTimeout),
		),
	}
	client.ctx, client.cancel = context.WithCancel(context.Background())

	return client
}

// login wipes olds credentials, logs in, and updates credentials if successful.
func (c *Client) login() error {
	c.credentials = new(LoginRes)

	if c.LoginReq.Username == "" || c.LoginReq.Password == "" {
		return errors.Errorf("invalid CLC credentials %q/%q", c.LoginReq.Username, c.LoginReq.Password)
	}

	if err := c.getCLCResponse("POST", "/v2/authentication/login", &c.LoginReq, c.credentials); err != nil {
		return err
	}
	c.AccountAlias = c.credentials.AccountAlias
	c.LocationAlias = c.credentials.LocationAlias

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
		jsonReq, err := json.Marshal(reqModel)
		if err != nil {
			return errors.Errorf("failed to encode request model %T %+v: %s", reqModel, reqModel, err)
		}
		reqBody = bytes.NewBuffer(jsonReq)
	}

	/* resModel must be a pointer type (call-by-value) */
	if resModel != nil {
		if resType := reflect.TypeOf(resModel); resType.Kind() != reflect.Ptr {
			return errors.Errorf("Expecting pointer to result model %T", resModel)
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

	if Debug && c.Log != nil {
		reqDump, _ := httputil.DumpRequest(req, true)
		c.Log.Printf("%s", reqDump)
	}

	res, err := c.requestor.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if Debug && c.Log != nil {
		resDump, _ := httputil.DumpResponse(res, true)
		c.Log.Printf("%s", resDump)
	}

	switch res.StatusCode {
	case 200, 201, 202, 204: /* OK / CREATED / ACCEPTED / NO CONTENT */
		if resModel != nil {
			if res.ContentLength == 0 {
				return errors.Errorf("Unable do populate %T result model, due to empty %q response",
					resModel, res.Status)
			}
			return json.NewDecoder(res.Body).Decode(resModel)
		} else if res.ContentLength > 0 {
			return errors.Errorf("Unable to decode non-empty %q response (%d bytes) to nil response model",
				res.Status, res.ContentLength)
		}
		return nil
	case 401:
		// This is returned if the BearerToken is m<issing or has become stale.
		if c.retryingLogin {
			return errors.New("failed to re-authenticate, credentials may be invalid")
		}
		if _, isLoginReq := reqModel.(*LoginReq); !isLoginReq {
			if Debug && c.Log != nil {
				c.Log.Printf("credentials are stale, retrying login ...")
			}
			// FIXME: the following is not thread-safe (multiple concurrent clients):
			c.retryingLogin = true
			if err = c.login(); err != nil {
				return err
			}
			if Debug && c.Log != nil {
				c.Log.Printf("re-authentication worked, retrying request ...")
			}
			if err = c.getResponse(url, verb, reqModel, resModel); err != nil {
				return err
			}
			c.retryingLogin = false
			return nil
		}
		return ErrCredentialsInValid
	}

	// Remaining error cases: res.ContentLength is not reliable - in the SBS case, it used
	// Transfer-Encoding "chunked", without a Content-Length.
	body, err := ioutil.ReadAll(res.Body)
	if err != nil && res.ContentLength > 0 {
		return errors.Errorf("failed to read error response %d body: %s", res.StatusCode, err)
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
		err = errors.Errorf("%s (status: %d)", errMsg, res.StatusCode)
	} else {
		err = errors.New(res.Status)
	}
	return
}
