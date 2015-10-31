package clcv2

import (
_	"time"
	"fmt"
)

type Server struct {
	// Server ID
	Id		string

	// Name of the server
	Name		string

	// User-defined description of this server
	Description	string

	// UUID of the parent hardware group
	GroupId		string

	// Whether this is a custom template or running server
	IsTemplate	bool

	// Data center that this server resides in
	LocationId	string

	// Friendly name of the Operating System the server is running
	OsType		string

	// Describes whether the server is active or not
	Status		string

	// Resource allocations, alert policies, snapshots, and more.
	Details		struct {
		// Details about IP addresses associated with the server
		IpAddresses		[]struct{
			// Private IP address.
			Internal	string

			// If applicable, the public IP
			// If associated with a public IP address, then the "public" value is populated
			Public		string
		}

		// Describe each alert policy applied to the server
		AlertPolicies		[]struct{
			// Unique identifier of the policy
			Id	string
			// User-defined name of the alert policy
			Name	string
			// Collection of entity links that point to resources related to this policy
			Links	[]Link
		}

		// How many vCPUs are allocated to the server
		Cpu			int

		// How many disks are attached to the server
		DiskCount		int

		// Fully qualified name of the server
		Hostname		string

		// Indicator of whether server has been placed in maintenance mode
		InMaintenanceMode	bool

		// How many MB of memory are allocated to the server
		MemoryMb		int

		// Whether the server is running or not
		PowerState		string

		// How many total GB of storage are allocated to the server
		StorageGb		int

		// The disks attached to the server
		Disks			[]struct{
			// Unique identifier of the disk
			Id		string
			// Size of the disk in GB
			SizeGb		int
			//  List of partition paths on the disk
			PartitionPaths	[]string
		}

		// The partitions defined for the server
		Partitions		[]struct{
			// Size of the partition in GB
			SizeGb	float64
			// File system location path of the partition
			Path	string
		}

		// Details about any snapshot associated with the server
		Snapshots		[]struct{
			// Timestamp of the snapshot
			// FIXME: maybe time.Time?
			Name	string

			// Collection of entity links that point to resources related to this snapshot
			Links	[]Link
		}

		// Details about any custom fields and their values
		CustomFields		[]CustomField

		// Processor configuration description (for bare metal servers only)
		ProcessorDescription	string

		// Storage configuration description (for bare metal servers only)
		StorageDescription	string
	}

	// Whether a standard or premium server
	Type		string

	// Whether it uses standard or premium storage
	StorageType	string

	// Describes "created" and "modified" details
	ChangeInfo	ChangeInfo

	// Collection of entity links that point to resources related to this server
	Links		[]Link
}

// Get the details for a individual server.
// @serverId: ID of the server being queried
func (c *Client) GetServer(serverId string) (res Server, err error) {
	err = c.getResponse("GET", fmt.Sprintf("/v2/servers/%s/%s", c.AccountAlias, serverId), nil, &res)
	return
}

type ServerStatus struct {
	// ID of the server that the operation was performed on.
	Server		string

	// Boolean indicating whether the operation was successfully added to the queue.
	IsQueued	bool

	// Collection of entity links that point to resources related to this server operation.
	Links		[]Link

	// If something goes wrong or the request is not queued,
	// this is the message that contains the details about what happened.
	ErrorMessage	string
}

// Send the delete operation to a given server and add operation to queue.
// @serverId: ID of the server to be deleted.
func (c *Client) DeleteServer(serverId string) (res ServerStatus, err error) {
	err = c.getResponse("DELETE", fmt.Sprintf("/v2/servers/%s/%s", c.AccountAlias, serverId), nil, &res)
	return
}

/*
 * OVF Import API
 */
type ImportOVF struct {
	// ID of the OVF.
	Id		string
	// Name of the OVF.
	Name		string
	// Number of GB of storage the server is configured with.
	StorageSizeGb	int
	// Number of processors the server is configured with.
	CpuCount	int
	// Number of MB of memory the server is configured with.
	MemorySizeMb	int
}

// Get the list of available servers that can be imported.
// @locationId: Data center location identifier
func (c *Client) GetServerImports(locationId string) (res []ImportOVF, err error) {
	err = c.getResponse("GET", fmt.Sprintf("/v2/vmImport/%s/%s/available", c.AccountAlias, locationId), nil, &res)
	return
}
