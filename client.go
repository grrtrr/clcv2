package clcv2

import (
	"bytes"
	"encoding/json"
	"errors"
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
)

const (
	// CenturyLink Cloud API url
	BaseURL = "https://api.ctl.io"

	// Maximum number of retries per request.
	MaxRetries = 3

	// Per-request retry delay for the retryer.
	StepDelay = time.Second * 10
)

// Errors returned by the Client
var (
	ErrUnauthencicated = errors.New("authentication error: credentials are stale or invalid.")
)

// Client wraps a http.Client, along with credentials and logging information.
type Client struct {
	requestor *http.Client

	// Authentication information
	LoginRes

	// Logger used for (debugging) output.
	Log logrus.StdLogger
}

// LoginRes is the data returned by the CLCv2 endpoint after successful authentication.
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

func (l LoginRes) String() string {
	return fmt.Sprintf("User=%s, Account=%s, Location=%s, Roles=%s", l.UserName,
		l.AccountAlias, strings.Join(l.Roles, ", "))
}

// NewClient initializes the parts common to both Client and CLIClient
func NewClient() *Client {
	client := &Client{}
	client.requestor = &http.Client{
		Transport: rehttp.NewTransport(nil, // default transport
			client.retryer(MaxRetries),
			// Note: using g_timeout as upper bound for the exponential backoff.
			//       This means g_timeout has to be large enough to run MaxRetries
			//       requests with individual retries.
			rehttp.ExpJitterDelay(StepDelay, g_timeout),
		),
	}
	client.SetTimeout(g_timeout)

	return client
}

// Log in and update credentials if successful. Requires c.LoginRes.BearerToken to be empty.
// @user: CLCv2 control portal username
// @pass: CLCv2 control portal password
func (c *Client) login(user, pass string) error {
	var LoginReq = struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{user, pass}
	return c.getResponse("POST", "/v2/authentication/login", &LoginReq, &c.LoginRes)
}

// retryer implements the retry policy: (a) any failure, (b) temporary failure status codes
func (c *Client) retryer(maxRetries int) rehttp.RetryFn {
	return rehttp.RetryFn(func(at rehttp.Attempt) bool {
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

// Perform a v2 API request
// @verb:     Http verb to use
// @path:     relative to BaseURL (includes the 'v2' version).
// @reqModel: request model to serialize, or nil.
// @resModel: result model to deserialize, must be a pointer to the expected result, or nil.
// Evaluates the StatusCode of the BaseResponse (embedded) in @inModel and sets @err accordingly.
// If @err == nil, fills in @resModel, else returns error.
func (c *Client) getResponse(verb, path string, reqModel, resModel interface{}) (err error) {
	var reqBody io.Reader
	fmt.Println("client base getResponse", verb, path)
	if reqModel != nil {
		if g_debug {
			c.Log.Printf("reqModel %T %+v\n", reqModel, reqModel)
		}

		jsonReq, err := json.Marshal(reqModel)
		if err != nil {
			return fmt.Errorf("Failed to encode request model %T %+v: %s", reqModel, reqModel, err)
		}
		reqBody = bytes.NewBuffer(jsonReq)
	}

	/* resModel must be a pointer type (call-by-value) */
	if resModel != nil {
		if resType := reflect.TypeOf(resModel); resType.Kind() != reflect.Ptr {
			return fmt.Errorf("Expecting pointer to result model %T", resModel)
		} else if g_debug {
			c.Log.Printf("resModel %T %+v", resModel, resModel)
		}
	}

	req, err := http.NewRequest(verb, BaseURL+path, reqBody)
	if err != nil {
		return
	}
	if c.BearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.BearerToken)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json")

	if g_debug {
		reqDump, _ := httputil.DumpRequest(req, true)
		c.Log.Printf("%s", reqDump)
	}

	res, err := c.requestor.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()

	if g_debug {
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
		return ErrUnauthencicated
	}

	/* Remaining error cases: */
	if res.ContentLength > 0 {
		var errMsg string
		var body []byte

		if body, err = ioutil.ReadAll(res.Body); err != nil {
			return fmt.Errorf("Failed to read error response %d body: %s", res.StatusCode, err)
		}

		// Currently 3 different types of response have been observed:
		// 1) bare JSON string
		// 2) struct { message: "string" }
		// 3) struct { message: "string", "modelState": map[string]interface{} }
		//    E.g.:  {"":["The server must be in Active or Archived state."]}
		//	      "modelState":{"body.networkId":["The network vlan_1249_10.81.149 is not valid."]}
		//	      "modelState":{"":["The server must be in Active or Archived state."]}
		//
		errMsg = string(body)
		if ct, _, _ := mime.ParseMediaType(res.Header.Get("Content-Type")); ct == "application/json" {
			/* Code thanks to & inspired by clc-go-cli */
			var payload map[string]interface{}

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
			}
		}
		err = fmt.Errorf("%s (status: %d)", errMsg, res.StatusCode)
	} else {
		err = errors.New(res.Status)
	}
	return
}
