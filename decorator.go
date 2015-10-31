/*
 * Decorator Pattern; after Tomas Senart's excellent GopherCon 2015
 * talk "Embrace the interface"
 */
package clcv2

import (
	"net/http/httputil"
	"net/http"
	"log"
)

// A Requestor sends http.Requests and returns http.Responses or errors in the case of failure
type Requestor interface {
	Do(*http.Request) (*http.Response, error)
}

// RequestorFunc is a function type that implements the Requestor interface.
type RequestorFunc func(*http.Request) (*http.Response, error)

func (f RequestorFunc) Do(r *http.Request) (*http.Response, error) {
	return f(r)
}

// Decorator Pattern: wrap Requestor with extra behaviour
type Decorator func (Requestor) Requestor

// Logging Decorator
func Logging(l *log.Logger) Decorator {
	return func(c Requestor) Requestor {
		return RequestorFunc(func(r *http.Request) (*http.Response, error) {
			l.Printf("%s: %s %s", r.UserAgent(), r.Method, r.URL)
			if g_debug {
				reqDump, _ := httputil.DumpRequest(r, true)
				l.Printf("%s", reqDump)
			}

			res, err := c.Do(r)

			if g_debug {
				resDump, _ := httputil.DumpResponse(res, true)
				l.Printf("%s", resDump)
			}
			return res, err
		})
	}
}

// Header returns a Decorator that adds the given HTTP header to every request done by a Requestor
func Header(name, value string) Decorator {
	return func(c Requestor) Requestor {
		return RequestorFunc(func(r *http.Request) (*http.Response, error) {
			r.Header.Set(name, value)
			return c.Do(r)
		})
	}
}

func Authorization(token string) Decorator {
	return Header("Authorization", token)
}
