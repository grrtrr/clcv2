package clcv2

import (
	"fmt"
	"time"
)

// QueueStatus reflects the CLCv2 status according to https://www.ctl.io/api-docs/v2/#get-status#response
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
// @statusId: queue ID to query (contains location ID in the format of "wa1-<number>")
func (c *Client) GetStatus(statusId string) (status QueueStatus, err error) {
	path := fmt.Sprintf("/v2/operations/%s/status/%s", c.AccountAlias, statusId)
	err = c.getCLCResponse("GET", path, nil, &struct{ Status *QueueStatus }{&status})
	return
}

// PollStatus polls the queue status of @statusId until it reaches either %Succeeded or %Failed.
// @statusId:     queue ID to query
// @pollInterval: wait interval between poll attemps, use 0 for one-shot operation
func (c *Client) PollStatus(statusId string, pollInterval time.Duration) error {
	var prevStatus QueueStatus = Unknown
	for {
		status, err := c.GetStatus(statusId)
		if err != nil {
			return fmt.Errorf("Failed to query status of status ID %d: %s", statusId, err)
		}
		if status != prevStatus {
			fmt.Printf("%s %s: %s\n", time.Now().Format("15:04:05"), statusId, status)
			prevStatus = status
		}
		if pollInterval == 0 || status == Succeeded || status == Failed {
			break
		}
		time.Sleep(pollInterval)
	}
	return nil
}

// AwaitCompletion waits until @statusId completes.
// @statusId: queue ID to query
func (c *Client) AwaitCompletion(statusId string) (QueueStatus, error) {
	for {
		status, err := c.GetStatus(statusId)
		if err != nil {
			return Unknown, fmt.Errorf("unable to query status of %s: %s", statusId, err)
		}
		if status == Succeeded || status == Failed {
			return status, nil
		}
		time.Sleep(1 * time.Second)
	}
}

// Status struct returned by operations such as 'Delete Group' and similar.
type StatusLink struct {
	// The identifier of the job in queue.
	// Can be passed to Get Status call to retrieve status of job.
	Id string

	// The Link type (should be set to "status")
	Rel string

	// The URI for the 'Get Status' call for this resource
	Href string
}

// Like getCLCResponse, but extract the Status Id from the Links array contained in the response.
// Accordingly, since only the status Id is returned, this function does not take a @resModel.
func (c *Client) getStatus(verb, path string, reqModel interface{}) (statusId string, err error) {
	var sl StatusLink

	if err = c.getCLCResponse(verb, path, reqModel, &sl); err == nil {
		if sl.Rel != "status" {
			err = fmt.Errorf("Link information Rel-type not set to 'status' in %+v", sl)
		} else {
			statusId = sl.Id
		}
	}
	return
}

// StatusResponse is the type of response returned by operations such as
// CreateServer, CloneServer, DeleteServer, ImportServer,
// ArchiveServer, CreateSnapshot, ExecutePackage
type StatusResponse struct {
	// ID of the server that the operation was performed on.
	Server string

	// Boolean indicating whether the operation was successfully added to the queue.
	IsQueued bool

	// Collection of entity links that point to resources related to this server operation.
	Links []Link

	// If something goes wrong or the request is not queued,
	// this is the message that contains the details about what happened.
	ErrorMessage string
}

// Run an Http request and evaluate the returned %StatusResponse, return links
// @verb, @path, @reqModel: as in getCLCResponse()
// @useArray:               whether to expect a singleton StatusResponse, or an array with one such element
func (c *Client) getStatusResponse(verb, path string, useArray bool, reqModel interface{}) (res StatusResponse, err error) {
	if useArray {
		var status []StatusResponse

		if err = c.getCLCResponse(verb, path, reqModel, &status); err != nil {
			return
		} else if len(status) == 0 {
			err = fmt.Errorf("empty status response from server")
		} else if len(status) != 1 {
			err = fmt.Errorf("multiple status responses (%d) from server", len(status))
		} else {
			res = status[0]
		}
	} else {
		err = c.getCLCResponse(verb, path, reqModel, &res)
	}

	if err == nil {
		if res.ErrorMessage != "" {
			err = fmt.Errorf("request on %s failed - %s", res.Server, res.ErrorMessage)
		} else if !res.IsQueued {
			err = fmt.Errorf("request on %s was not queued", res.Server)
		}
	}
	return
}

// Wrap getStatusResponse() to only extract the statusId contained in the 'status' link
// @verb, @path, @useArray, @reqModel: as in getStatusResponse
func (c *Client) getStatusResponseId(verb, path string, useArray bool, reqModel interface{}) (statusId string, err error) {
	var status StatusResponse
	var link *Link

	status, err = c.getStatusResponse(verb, path, useArray, reqModel)
	if err != nil {
		return
	}
	if link, err = extractLink(status.Links, "status"); err == nil {
		statusId = link.Id
	}
	return
}
