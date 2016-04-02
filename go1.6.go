// +build go1.6

package clcv2

import "time"

// SetTimeout changes the per-request timeout.
// The client timeout applies to the request as a whole, i.e. including any retries.
// See http://0value.com/Let-the-Doer-Do-it
func (c *Client) SetTimeout(timeout time.Duration) {
	c.requestor.Timeout = timeout
}
