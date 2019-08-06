package clcv2

/*
 * Load Balancers as a Service (LBaas)
 * https://www.ctl.io/api-docs/v2/#lbaas
 */

import (
	"fmt"
	"time"

	uuid "github.com/satori/go.uuid"
)

// This service uses a different API endpoint
const lbaasBaseUrl = "https://api.loadbalancer.ctl.io"

// LbCreateRequest represents CLCv2 load balancer data
type LbCreateRequest struct {
	// ID of the 'Create LBaaS' request
	ID uuid.UUID `json:"id"`

	// Describes the activity within the 'Create LBaaS' request
	Description string `json:"description"`

	// Status of the load balancer ("ACTIVE", "DELETED")
	Status string `json:"status"`

	// Time of the request to create the load balancer instance
	Created LbEpochSeconds `json:"requestDate"`

	// Time the request completed (null until then)
	Completed *LbEpochSeconds `json:"completionDate"`

	// Collection of entity links that point to resources related to this load balancer
	Links []Link `json:"links"`
}

// CreateLbInstance creates a new load balancer in @dc.
// @name:   name of the load balancer to create
// @desc:   textual description of the load balancer
// @dc:     location alias of the data centre in which to create the load balancer
func (c *Client) CreateLbInstance(name, desc, dc string) (req LbCreateRequest, err error) {
	var path = fmt.Sprintf("/%s/%s/loadbalancers", c.AccountAlias, dc)

	return req, c.getLbResponse("POST", path, struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}{name, desc}, &req)
}

// GetLbCreateRequest retrieves the current status of the 'Create LB instance' request @id.
// @dc: location alias of the data centre associated with @id
// @id: LBaaS instance UUID
func (c *Client) GetLbCreateRequest(dc, id string) (req LbCreateRequest, err error) {
	var path = fmt.Sprintf("/%s/%s/loadbalancers/requests/%s", c.AccountAlias, dc, id)

	return req, c.getLbResponse("GET", path, nil, &req)
}

// LbInstance represents an LBaaS instance
type LbInstance struct {
	// UUID of the Load Balancer
	ID uuid.UUID

	// Name of this load balancer instance
	Name string

	// A short description of this load balancer
	Description string

	// The external (public) IP address of the load balancer
	PublicIP string `json:"publicIPAddress"`

	// Collection of pools configured for this load balancer
	// pools

	// Status of the load balancer: active, deleted, creating, or failed
	Status string

	// The account which owns the load balancer
	Owner string `json:"accountAlias"`

	// The data center in which the load balancer resides
	DataCenter string

	// Date-time stamp of the load balancer creation
	Created LbEpochSeconds `json:"creationTime"`

	// Time of the load balancer deletion. Will be null if load balancer not in deleted status.
	Deleted *LbEpochSeconds `json:"deletionTime"`

	// This seems to be something for internal use:
	KeepAliveRouter string `json:"keepalivedRouterId,omitempty"`
}

// GetLbInstances returns the list all LBaaS instances in the data center @dc.
func (c *Client) GetLbInstances(dc string) ([]LbInstance, error) {
	var path = fmt.Sprintf("/%s/%s/loadbalancers", c.AccountAlias, dc)
	var result struct {
		Values []LbInstance
	}

	return result.Values, c.getLbResponse("GET", path, nil, &result)
}

// DeleteLbInstance deletes the load balancer @id in @dc
// @id: UUID of the load balancer to delete
// @dc: location alias of the data centre the load balancer resides in
func (c *Client) DeleteLbInstance(id, dc string) error {
	var path = fmt.Sprintf("/%s/%s/loadbalancers/%s", c.AccountAlias, dc, id)
	return c.getLbResponse("DELETE", path, nil, nil)
}

// LbEpochSeconds is the custom date/time format used by the LBaaS API.
type LbEpochSeconds uint64

func (l LbEpochSeconds) Time() time.Time {
	return time.Unix(int64(l)/1000, int64(l)%1000)
}

func (l LbEpochSeconds) String() string {
	return l.Time().String()
}

// getLbResponse is like getCLCResponse but hits the LBaaS API base URL instead.
func (c *Client) getLbResponse(verb, path string, reqModel, resModel interface{}) error {
	return c.getResponse(lbaasBaseUrl+path, verb, reqModel, resModel)
}
