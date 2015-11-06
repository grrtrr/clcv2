package clcv2

import (
	"net/http/httputil"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"
	"errors"
	"bytes"
	"time"
	"mime"
	"flag"
	"log"
	"fmt"
	"os"
	"io"
)

const (
	BaseURL = "https://api.ctl.io"
)

/* Global variables */
var (
	g_user, g_pass	string		/* Command-line username/password */
	g_timeout	time.Duration	/* Client default timeout */
	g_debug		bool		/* Command-line debug flag */
)

func init() {
	flag.BoolVar(&g_debug,  "d", false, "Produce debug output")
	flag.StringVar(&g_user, "u", "",    "CLC Login Username")
	flag.StringVar(&g_pass, "p", "",    "CLC Login Password")
	/*
	  * Caveat: keep the timeout value high, at least a few minutes.
	  *         Some operations, such as querying details of a new server immediately
	  *         after launching a CreateServer request, can take up to circa a minute.
	  */
	flag.DurationVar(&g_timeout, "timeout", 180 * time.Second, "Client default timeout")
}

type Client struct {
	requestor	Requestor

	// Authentication information
	*LoginRes

	// Logger to use by this package
	 Log        *log.Logger
}

// Return authenticated client
// The environment variable CLC_ALIAS, if set, is allowed to override the default LocationAlias.
func NewClient() (client *Client, err error) {
	client = &Client{ requestor: &http.Client{Timeout: g_timeout} }

	if g_debug {
		client.Log = log.New(os.Stdout, "", log.Ltime | log.Lshortfile)
	} else {
		client.Log = log.New(ioutil.Discard, "", 0)
	}

	if err = client.loadCredentials(); err != nil {
		return
	}

	if alias := os.Getenv("CLC_ALIAS"); alias != "" {
		client.LoginRes.LocationAlias = alias
	}
	client.requestor = Authorization("Bearer " + client.BearerToken)(client.requestor)
	return client, nil
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
			c.Log.Printf("resModel %T %+v\n", resModel, resModel)
		}
	}

	req, err := http.NewRequest(verb, BaseURL + path, reqBody)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept",       "application/json")

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
	case 200, 201, 202, 204:	/* OK / CREATED / ACCEPTED / NO CONTENT */
		if resModel != nil {
			return json.NewDecoder(res.Body).Decode(resModel)
		} else if res.ContentLength > 0 {
			return fmt.Errorf("Unable to decode non-empty response (%d bytes) to nil response model",
					  res.ContentLength)
		}
		return nil
	}

	if res.ContentLength > 0 {
		var errMsg string
		var body []byte

		if body, err = ioutil.ReadAll(res.Body); err != nil {
			return fmt.Errorf("Failed to read error response %d body: %s", res.StatusCode, err)
		}

		/*
		 * Currently 3 different types of response have been observed:
		 * 1) bare JSON string
		 * 2) struct { message: "string" }
		 * 3) struct { message: "string", "modelState": map[string]interface{} }
		      E.g.:  {"":["The server must be in Active or Archived state."]}
		              "modelState":{"body.networkId":["The network vlan_1249_10.81.149 is not valid."]}
		              "modelState":{"":["The server must be in Active or Archived state."]}
		 */
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
