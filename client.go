package clcv2

import (
	"net/http/httputil"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"
	"errors"
	"bytes"
	"mime"
	"log"
	"fmt"
	"os"
	"io"
)

const (
	BaseURL = "https://api.ctl.io"
)

type Client struct {
	requestor	Requestor

	// Authentication information
	*LoginRes

	// Logger to use by this package
	 Log        *log.Logger
}

/* Return authenticated client */
func NewClient(creds *LoginRes, logger *log.Logger) (*Client, error) {
	if g_debug {
		logger = log.New(os.Stdout, "", log.Ltime | log.Lshortfile)
	} else if logger == nil {
		logger = log.New(ioutil.Discard, "", log.Ltime | log.Lshortfile)
	}

	client := &Client{
		requestor: &http.Client{Timeout: g_timeout},
		LoginRes:  creds,
		Log:       logger,
	}

	if creds == nil {
		if err := client.login(); err != nil {
			return nil, err
		}
	}
	client.requestor = Authorization("Bearer " + client.BearerToken)(client.requestor)
	return client, nil
}

// Perform a v2 API request
// @verb:     Http verb to use
// @path:     relative to BaseURL (excludes the 'v2' version)
// @reqModel: request model to serialize, or nil
// @resModel: result model to deserialize, must be a pointer to the expected result
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
	if resModel == nil {
		return fmt.Errorf("Result model can not be nil")
	} else if resType := reflect.TypeOf(resModel); resType.Kind() != reflect.Ptr {
		return fmt.Errorf("Expecting pointer to result model %T", resModel)
	} else if g_debug {
		c.Log.Printf("resModel %T %+v\n", resModel, resModel)
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
		return json.NewDecoder(res.Body).Decode(resModel)
	}

	if res.ContentLength > 0 {
		var errMsg string
		var body []byte

		if body, err = ioutil.ReadAll(res.Body); err != nil {
			return fmt.Errorf("Failed to read error response %d body: %s", res.StatusCode, err)
		}

		if ct, _, _ := mime.ParseMediaType(res.Header.Get("Content-Type")); ct != "application/json" {
			errMsg = string(body)
		} else if err = json.Unmarshal(body, &struct { Message *string } { &errMsg }); err != nil {
			/* Not a { message: "string" }, variant, try bare string next */
			if err = json.Unmarshal(body, &errMsg); err != nil {
				errMsg = string(body)
			}
		}
		err = fmt.Errorf("%s (status: %d)", errMsg, res.StatusCode)
	} else {
		err = errors.New(res.Status)
	}
	return
}
