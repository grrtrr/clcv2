package clcv2

import "fmt"

type QueueStatus string

const (
	NotStarted QueueStatus = "notStarted"
	Executing              = "executing"
	Succeeded              = "succeeded"
	Failed                 = "failed"
	Resumed                = "resumed"
	Unknown                = "unknown"
)

// Get the status of a particular job in the queue, which keeps track of any long-running
// asynchronous requests (such as server power operations or provisioning tasks).
// Use this API operation when you want to check the status of a specific job in the queue.
// It is usually called after running a batch job and receiving the job identifier from the
// status link (e.g. Power On Server, Create Server, etc.) and will typically continue
// to get called until a "succeeded" or "failed" response is returned.
// @statusId: ID of the server being queried
func (c *Client) GetStatus(statusId string) (status QueueStatus, err error) {
	path := fmt.Sprintf("/v2/operations/%s/status/%s", c.AccountAlias, statusId)
	err = c.getResponse("GET", path, nil, &struct{ Status *QueueStatus } { &status })
	return
}
