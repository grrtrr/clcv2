package clcv2

/*
 * Load Balancers
 */

import "fmt"

// LoadBalancerStatus reflects the current state of a shared load balancer
type LoadBalancerStatus string

const (
	LbEnabled  LoadBalancerStatus = "enabled"
	LbDisabled                    = "disabled"
	LbDeleted                     = "deleted"
)

// LoadBalancer represents CLCv2 load balancer data
type LoadBalancer struct {
	// ID of the load balancer
	ID string

	// Friendly :) name of the load balancer
	Name string

	// Description for the load balancer
	Description string

	// The external (public) IP address of the load balancer
	IpAddress string

	// Status of the load balancer
	Status LoadBalancerStatus

	// Collection of pools configured for this shared load balancer
	Pools []LoadBalancerPool

	// Collection of entity links that point to resources related to this load balancer
	Links []Link
}

// LoadBalancerPool represents an individual load-balancer pool in a datacenter
type LoadBalancerPool struct {
	// ID of the load balancer pool
	ID string

	// Port configured on the public-facing side of the load balancer pool.
	Port int

	// The balancing method for this load balancer, either "leastConnection" or "roundRobin".
	Method string

	// The persistence method for this load balancer, either "standard" or "sticky".
	Persistence string

	// Collection of nodes configured behind this shared load balancer
	Nodes []LoadBalancerNode

	// Collection of entity links that point to resources related to this load balancer pool
	Links []Link
}

// LoadBalancerNode represents a node within a LoadBalancerPool
type LoadBalancerNode struct {
	// Name of the node (generally the IP address)
	Name string

	// Status of the node
	Status LoadBalancerStatus

	// The internal (private) IP address of the node server
	IPAddress string

	// The internal (private) port of the node server
	PrivatePort int

	// Collection of entity links that point to resources related to this node
	Links []Link
}

// GetSharedLoadBalancers returns the list of shared load balancers for a given account and data center.
// @dc: location alias of data centre to query
func (c *Client) GetSharedLoadBalancers(dc string) (lb []LoadBalancer, err error) {
	path := fmt.Sprintf("/v2/sharedLoadBalancers/%s/%s", c.AccountAlias, dc)
	err = c.getCLCResponse("GET", path, nil, &lb)
	return lb, err
}

// CreateSharedLoadBalancer creates a new load balancer in @dc in @active state.
// @name:   name of the load balancer to create
// @desc:   textual description of the load balancer
// @active: whether to create the load balancer in 'enabled' state, one of "enabled" or "disabled"
// @dc:     location alias of the data centre in which to create the load balancer
func (c *Client) CreateSharedLoadBalancer(name, desc, active, dc string) (lb LoadBalancer, err error) {
	path := fmt.Sprintf("/v2/sharedLoadBalancers/%s/%s", c.AccountAlias, dc)
	err = c.getCLCResponse("POST", path, struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Status      string `json:"status"`
	}{name, desc, active}, &lb)
	return lb, err
}

// DeleteSharedLoadbalancer deletes the load balancer @id in @dc
// @id: ID of the load balancer to delete
// @dc: location alias of the data centre the load balancer resides in
func (c *Client) DeleteSharedLoadBalancer(id, dc string) error {
	path := fmt.Sprintf("/v2/sharedLoadBalancers/%s/%s/%s", c.AccountAlias, dc, id)
	return c.getCLCResponse("DELETE", path, nil, nil)
}
