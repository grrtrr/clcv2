// +build !go1.6

package clcv2

import "time"

// SetTimeout changes the per-request timeout.
func (c *Client) SetTimeout(timeout time.Duration) {
	// Using non-0 timeout conflicts with the CancelRequest functionality
	// required by the rehttp package. This is fixed in go >= 1.6
	c.requestor.Timeout = 0
}
