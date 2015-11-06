package clcv2

import (
	"time"
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

// Query Server details by URI path.
// @path: relative path of the server, as e.g. returned via 'self' link in CreateServer
func (c *Client) GetServerByURI(path string) (res Server, err error) {
	err = c.getResponse("GET", path, nil, &res)
	return
}

// Get the details for a individual server.
// @serverId: ID/name of the server being queried (e.g. WA1DTGDFEDAD0
func (c *Client) GetServer(serverId string) (res Server, err error) {
	return c.GetServerByURI(fmt.Sprintf("/v2/servers/%s/%s", c.AccountAlias, serverId))
}


/*
 * Create Server
 */
type CreateServerReq struct {
	// Name of the server to create. Alphanumeric characters and dashes only.
	// Must be between 1-8 characters depending on the length of the account alias.
	// The combination of account alias and server name here must be no more than 10 characters in length.
	// (This name will be appended with a two digit number and prepended with the datacenter code
	// and account alias to make up the final server name.)
	Name			string			`json:"name"`

	// User-defined description of this server
	Description		string			`json:"description,omitempty"`

	// ID of the parent group. Retrieved from query to parent group, or by looking at the URL on the UI pages in the Control Portal.
	GroupId			string			`json:"groupId"`

	// ID of the server to use a source. May be the ID of a template, or when cloning, an existing server ID.
	SourceServerId		string			`json:"sourceServerId"`

	// Whether to create the server as managed or not. Default is false. (Ignored for bare metal servers.)
	IsManagedOs		bool			`json:"isManagedOS"`

	// Whether to add managed backup to the server. Must be a managed OS server. (Ignored for bare metal servers.)
	IsManagedBackup		bool			`json:"isManagedBackup"`

	// Primary DNS to set on the server. If not supplied the default value set on the account will be used.
	PrimaryDns		string			`json:"primaryDns,omitempty"`

	// Secondary DNS to set on the server. If not supplied the default value set on the account will be used.
	SecondaryDns		string			`json:"secondaryDns,omitempty"`

	// ID of the network to which to deploy the server. If not provided, a network will be chosen automatically.
	// If your account has not yet been assigned a network, leave this blank and one will be assigned automatically.
	NetworkId		string			`json:"networkId,omitempty"`

	// IP address to assign to the server. If not provided, one will be assigned automatically.
	// (Ignored for bare metal servers.)
	IpAddress		string			`json:"ipAddress,omitempty"`

	// Password of administrator or root user on server. If not provided, one will be generated automatically.
	Password		string			`json:"password,omitempty"`

	// Password of the source server, used only when creating a clone from an existing server.
	// (Ignored for bare metal servers.)
	SourceServerPassword	string			`json:"sourceServerPassword,omitempty"`

	// Number of processors to configure the server with (1-16) (ignored for bare metal servers)
	Cpu			int			`json:"cpu"`

	// ID of the vertical CPU Autoscale policy to associate the server with. (Ignored for bare metal servers.)
	CpuAutoscalePolicyId	string			`json:"cpuAutoscalePolicyId,omitempty"`

	// Number of GB of memory to configure the server with (1-128) (ignored for bare metal servers)
	MemoryGB		int			`json:"memoryGB"`

	// Whether to create a 'standard', 'hyperscale', or 'bareMetal' server
	Type			string			`json:"type"`

	// For standard servers, whether to use standard or premium storage.  (Ignored for bare metal servers.)
	// If not provided, will default to premium storage.
	StorageType		string			`json:"storageType,omitempty"`

	// ID of the Anti-Affinity policy to associate the server with. Only valid for hyperscale servers.
	AntiAffinityPolicyId	string			`json:"antiAffinityPolicyId,omitempty"`

	// Collection of custom field ID-value pairs to set for the server.
	CustomFields		[]SimpleCustomField	`json:"customFields"`

	// Collection of disk parameters (ignored for bare metal servers)
	AdditionalDisks		[]ServerDisk		`json:"additionalDisks"`

	// Date/time that the server should be deleted (ignored for bare metal servers)
	Ttl			time.Time		`json:"ttl"`

	// Collection of packages to run on the server after it has been built (ignored for bare metal servers)
	Packages		[]struct {
		// ID of the package to run on the server after it builds.
		PackageId	string	`json:"packageId"`

		// Collection of name-value pairs to specify package-specific parameters.
		Parameters	struct {
			name, value string	// FIXME lack of API documentation here
		}			`json:"parameters"`
	}						`json:"packages"`

	// Specifies the identifier for the specific configuration type of bare metal server to deploy.
	// Only required for bare metal servers. (Ignored for standard and hyperscale servers.)
	ConfigurationId		string			`json:"configurationId,omitempty"`

	// Specifies the OS to provision with the bare metal server.
	// Only required for bare metal servers. (Ignored for standard and hyperscale servers.)
	// Currently (Nov 3/2015), the only supported OS types are:
	// - redHat6_64Bit,
	// - centOS6_64Bit,
	// - windows2012R2Standard_64Bit,
	// - windows2012R2Datacenter_64Bit,
	// - ubuntu14_64Bit.
	OsType			string			`json:"osType"`
}

type ServerDisk struct {
	// File system path for disk (Windows drive letter or Linux mount point).
	// Must not be one of reserved names.
	Path	string	`json:"path"`

	// Amount in GB to allocate for disk, up to 1024 GB
	SizeGB	int	`json:"sizeGB"`

	// Whether the disk should be raw or partitioned
	Type	string	`json:"type"`
}


/* Status response, used by: CreateServer, CloneServer, DeleteServer, ImportServer,
                             ArchiveServer, CreateSnapshot, ExecutePackage,  */
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

// Run an Http request and evaluate the returned %ServerStatus, return links
// @verb, @path, @reqModel: as in getResponse()
func (c *Client) getServerStatus(verb, path string, reqModel interface{}) (res ServerStatus, err error) {
	if err = c.getResponse(verb, path, reqModel, &res); err == nil {
		if res.ErrorMessage != "" {
			err = fmt.Errorf("Request on %s failed: %s", res.Server, res.ErrorMessage)
		} else if !res.IsQueued {
			err = fmt.Errorf("Request on %s was not queued", res.Server)
		}
	}
	return
}

// Create a new server.
// @serverId: ID of the server to be deleted.
// Returns new server @name and @statusId if successful.
func (c *Client) CreateServer(req *CreateServerReq) (name, statusId string, err error) {
	var status ServerStatus
	var server Server
	var link *Link

	status, err = c.getServerStatus("POST", fmt.Sprintf("/v2/servers/%s", c.AccountAlias), req)
	if err != nil {
		return
	}

	if link, err = extractLink(status.Links, "status"); err != nil {
		return
	}
	statusId = link.Id

	if link, err = extractLink(status.Links, "self"); err != nil {
		return
	}

	/* Note: the following call can take long, at least circa a minute. Use appropriate timeout. */
	if server, err = c.GetServerByURI(link.Href); err != nil {
		err = fmt.Errorf("Failed to query details of server %s: %s", name, err)
	}
	name = server.Name
	return
}

// Send the delete operation to a given server and add operation to queue.
// @serverId: ID of the server to be deleted.
func (c *Client) DeleteServer(serverId string) (statusId string, err error) {
	var status ServerStatus
	var link *Link

	status, err = c.getServerStatus("DELETE", fmt.Sprintf("/v2/servers/%s/%s", c.AccountAlias, serverId), nil)
	if err != nil {
		return
	}
	if link, err = extractLink(status.Links, "status"); err == nil {
		statusId = link.Id
	}
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

/*
 * Credentials
 */
type ServerCredentials struct {
	// The username of root/administrator on the server.
	// Typically "root" for Linux machines and "Administrator" for Windows.
	Username        string

	// The administrator/root password used to login.
	Password        string
}

// Retrieve the administrator/root password on an existing server.
// @serverId: ID of the server with the credentials to return.
func (c *Client) GetServerCredentials(serverId string) (res ServerCredentials, err error) {
	err = c.getResponse("GET", fmt.Sprintf("/v2/servers/%s/%s/credentials", c.AccountAlias, serverId), nil, &res)
	return
}

// Change the administrator/root password on an existing server given the current administrator/root password.
// @serverId: ID of the server to change.
// @curPass:  current password for @serverId
// @newPass:  new password for @serverId
func (c *Client) ServerChangePassword(serverId, curPass, newPass string) (statusId string, err error) {
	var op = PatchOperation{
		Op:     "set",
		Member: "password",
		Value:  struct{
			// The current administrator/root password used to login.
			Current		string	`json:"current"`

			// The new administrator/root password to change to.
			Password	string	`json:"password"`
		} { curPass, newPass },
	}
	return c.patchStatus(fmt.Sprintf("/v2/servers/%s/%s", c.AccountAlias, serverId), &op)
}

// Change the number of CPU cores on an existing server.
// @serverId: ID of the server to change.
// @cpus:     number of CPUs to allocate for @serverId.
func (c *Client) ServerSetCpus(serverId, cpus string) (statusId string, err error) {
	return c.patchStatus(fmt.Sprintf("/v2/servers/%s/%s", c.AccountAlias, serverId),
			     &PatchOperation{ "set", "cpu", cpus })
}

// Change the amount of memory on an existing server.
// @serverId: ID of the server to change.
// @memGB:    amount of memory (in GB) to allocate.
func (c *Client) ServerSetMemory(serverId, memGB string) (statusId string, err error) {
	return c.patchStatus(fmt.Sprintf("/v2/servers/%s/%s", c.AccountAlias, serverId),
			     &PatchOperation{ "set", "memory", memGB })
}

// Change the description of an existing server.
// @serverId: ID of the server to change.
// @desc:     new description to use for @serverId.
func (c *Client) ServerSetDescription(serverId, desc string) error {
	return c.patch(fmt.Sprintf("/v2/servers/%s/%s", c.AccountAlias, serverId),
		       &PatchOperation{ "set", "description", desc })
}

// Change the description of an existing server.
// @serverId:   ID of the server to change.
// @parentUUID: UUID of new parent group for @serverId.
func (c *Client) ServerSetGroup(serverId, parentUUID string) error {
	return c.patch(fmt.Sprintf("/v2/servers/%s/%s", c.AccountAlias, serverId),
		       &PatchOperation{ "set", "groupId", parentUUID })
}
