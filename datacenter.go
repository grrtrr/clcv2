package clcv2

import (
	"fmt"
)

// DataCenter represents the information about a single data centre.
type DataCenter struct {
	// Short value representing the data center code
	Id string

	// Full, friendly name of the data center
	Name string

	// Collection of entity links that point to resources related to this data center
	Links []Link
}

// Get the list of data centers that a given account has access to.
func (c *Client) GetLocations() (loc []DataCenter, err error) {
	err = c.getCLCResponse("GET", "/v2/datacenters/"+c.AccountAlias, nil, &loc)
	return
}

// Get the details of a specific data center.
// @location:   location alias of data centre to query
// @groupLinks: whether to include 'group' type of links
func (c *Client) GetDatacenter(location string, groupLinks bool) (res DataCenter, err error) {
	path := fmt.Sprintf("/v2/datacenters/%s/%s?groupLinks=%t", c.credentials.AccountAlias, location, groupLinks)
	err = c.getCLCResponse("GET", path, nil, &res)
	return
}

// IntResourceValue represents an integer value.
type IntResourceValue struct {
	Value     uint64 `json:"value"`
	Inherited bool   `json:"inherited"`
}

// ComputeLimits represents the computational resource limits of a data centre.
type ComputeLimits struct {
	CPU       IntResourceValue `json:"cpu"`
	MemoryGB  IntResourceValue `json:"memoryGB"`
	StorageGB IntResourceValue `json:"storageGB"`
}

// Get the compute limits for the given data centre.
func (c *Client) GetDatacenterComputeLimits(location string) (*ComputeLimits, error) {
	path := fmt.Sprintf("/v2/datacenters/%s/%s/computeLimits", c.credentials.AccountAlias, location)
	res := new(ComputeLimits)
	return res, c.getCLCResponse("GET", path, nil, res)
}

// NetLimits represents the maximum number of networks allowed in a data centre.
type NetLimits struct {
	Networks IntResourceValue `json:"networks"`
}

// Get the networking limits for the given data centre.
func (c *Client) GetDatacenterNetworkLimits(location string) (*NetLimits, error) {
	path := fmt.Sprintf("/v2/datacenters/%s/%s/networkLimits", c.credentials.AccountAlias, location)
	res := new(NetLimits)
	return res, c.getCLCResponse("GET", path, nil, res)
}

/*
 * Deployment Capabilities
 */
type DeploymentCapabilities struct {
	// Whether or not this data center provides support for servers with premium storage
	SupportsPremiumStorage bool

	// Whether or not this data center provides support for shared load balancer configuration
	SupportsSharedLoadBalancer bool

	// Whether or not this data center provides support for provisioning bare metal servers
	SupportsBareMetalServers bool

	// FIXME: the following appear in the output, but are not documented
	DataCenterEnabled bool
	ImportVMEnabled   bool

	// Collection of networks that can be used for deploying servers
	DeployableNetworks []struct {
		// User-defined name of the network
		Name string

		// Unique identifier of the network
		NetworkId string

		// Network type, usually "private" for networks created by the user
		Type string

		// Account alias for the account in which the network exists
		AccountID string
	}

	// Collection of available templates in the data center that can be used to create servers
	Templates []struct {
		// Underlying unique name for the template
		Name string

		// Description of the template at it appears in the Control Portal UI
		Description string

		// FIXME: the following appears in the output, but is not documented
		OsType string

		// The amount of storage allocated for the primary OS root drive
		StorageSizeGB int

		// List of capabilities supported by this specific OS template
		// (example: whether adding CPU or memory requires a reboot or not)
		Capabilities []string

		// List of drive path names reserved by the OS that can't be used to name user-defined drives
		ReservedDrivePaths []string

		// Length of the string for naming a drive path, if applicable
		DrivePathLength int
	}

	// Collection of available OS types that can be imported as virtual machines.
	ImportableOsTypes []struct {
		// FIXME: no online description for this as yet
		Id                 int
		Type               string
		Description        string
		LabProductCode     string
		PremiumProductCode string
	}
}

// Get the list of capabilities that a specific data center supports for a given account,
// including the deployable networks, OS templates, and whether features like premium storage
// and shared load balancer configuration are available.
// @location:   location alias of data centre to query
func (c *Client) GetDeploymentCapabilities(location string) (res DeploymentCapabilities, err error) {
	path := fmt.Sprintf("/v2/datacenters/%s/%s/deploymentCapabilities", c.credentials.AccountAlias, location)
	err = c.getCLCResponse("GET", path, nil, &res)
	return
}

/*
 * Bare Metal Capabilities
 */
type BareMetalCapabilities struct {
	// Collection of available bare metal configuration types to pass in as
	// configurationId when creating a bare-metal server
	Skus []struct {
		// The configurationId to pass to the Create Server API operation when creating a bare metal server.
		Id string

		// Price per hour for the given configuration.
		HourlyRate float64

		// The level of availability for the given configuration: either high, low, or none.
		Availability string

		// Information about the memory on the server.
		memory []struct {
			// Memory capacity in gigabytes
			CapacityGB int
		}

		// Information about the physical processors on the server.
		Processor struct {
			// Description of the processor including model and clock speed
			Description string

			// Number of cores for each processor socket
			CoresPerSocket int

			// Number of sockets
			Sockets int
		}

		// Collection of disk information, each item representing one physical disk on the server.
		Storage []struct {
			// Underlying unique name for the OS type
			CapacityGB int

			// RPM (revolutions per minutes) speed of the disk
			SpeedRpm int

			// Disk type. Only Hdd currently supported.
			Type string
		}
	}

	// Collection of available operating systems when creating a bare metal server
	OperatingSystems []struct {
		// Underlying unique name for the OS type
		Type string

		// Friendly description for the OS type
		Description string

		// Price per hour per socket for the OS type.
		HourlyRatePerSocket float64
	}
}

// Get the list of bare metal capabilities that a specific data center supports for a given account,
// including the list of configuration types and the list of supported operating systems.
// @location:   location alias of data centre to query
func (c *Client) GetBareMetalCapabilities(location string) (res BareMetalCapabilities, err error) {
	path := fmt.Sprintf("/v2/datacenters/%s/%s/bareMetalCapabilities", c.credentials.AccountAlias, location)
	err = c.getCLCResponse("GET", path, nil, &res)
	return
}
