package clcv2

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
)

const (
	// CLC Password unsupported characters. FIXME: possibly subject to change without notice.
	InvalidPasswordCharacters = `"';^&<>|\`
)

// Various server-related errors returned by this package
var (
	// ErrNoSnapshot is returned when trying to delete a non-existing snapshot
	ErrNoSnapshot = errors.New("no snapshot exists")
)

// Server represents the CLC-specific server structure
type Server struct {
	// Server ID
	Id string

	// Name of the server
	Name string

	// User-defined description of this server
	Description string

	// UUID of the parent hardware group
	GroupId string

	// Whether this is a custom template or running server
	IsTemplate bool

	// Data center that this server resides in
	LocationId string

	// Friendly name of the Operating System the server is running
	OsType string

	// Describes whether the server is active: "underConstruction", "active", "queuedForDelete"
	Status string

	// Resource allocations, alert policies, snapshots, and more.
	Details struct {
		// Details about IP addresses associated with the server
		IpAddresses []ServerIPAddress

		// Describe each alert policy applied to the server
		AlertPolicies []struct {
			// Unique identifier of the policy
			Id string
			// User-defined name of the alert policy
			Name string
			// Collection of entity links that point to resources related to this policy
			Links []Link
		}

		// How many vCPUs are allocated to the server
		Cpu int

		// How many disks are attached to the server
		DiskCount int

		// Fully qualified name of the server
		Hostname string

		// Indicator of whether server has been placed in maintenance mode
		InMaintenanceMode bool

		// How many MB of memory are allocated to the server
		MemoryMb int

		// Whether the server is running or not: "started", "paused", or "stopped"
		PowerState string

		// How many total GB of storage are allocated to the server
		StorageGb int

		// The disks attached to the server
		Disks []struct {
			// Unique identifier of the disk
			Id DiskID

			// Size of the disk in GB
			SizeGB uint32

			// List of partition paths on the disk (seems to always be empty)
			PartitionPaths []string
		}

		// The partitions defined for the server
		Partitions []struct {
			// Size of the partition in GB
			SizeGB float64
			// File system location path of the partition
			Path string
		}

		// Details about any snapshot associated with the server
		Snapshots []ServerSnapshot

		// Details about any custom fields and their values
		CustomFields []CustomField

		// Processor configuration description (for bare metal servers only)
		ProcessorDescription string

		// Storage configuration description (for bare metal servers only)
		StorageDescription string
	}

	// Whether a standard or premium server
	Type string

	// Whether it uses standard or premium storage
	StorageType string

	// Describes "created" and "modified" details
	ChangeInfo ChangeInfo

	// Collection of entity links that point to resources related to this server
	Links []Link
}

// IPs returns the list of IPs of @s as flattened list
func (s *Server) IPs() []string {
	var seen = make(map[string]bool) /* track private IPs, they can be mapped to public ones */
	var ips []string

	if s != nil {
		for _, ip := range s.Details.IpAddresses {
			if ip.Public != "" {
				ips = append(ips, ip.Public)
			}
			if !seen[ip.Internal] {
				ips = append(ips, ip.Internal)
				seen[ip.Internal] = true
			}
		}
	}
	return ips
}

// ServerIPAddress represents an IP address attached to a server.
type ServerIPAddress struct {
	// Private IP address. If @Public is non-empty, @Internal is associated with @Public.
	Internal string

	// If applicable, the public IP (if empty, @Internal is a private IP address).
	Public string
}

// IsPublic returns true if @s is a public ServerIPAddress
func (s *ServerIPAddress) IsPublic() bool {
	return s.Public != ""
}

// Query Server details by URI path.
// @path: relative path of the server, as e.g. returned via 'self' link in CreateServer
func (c *Client) GetServerByURI(path string) (res Server, err error) {
	err = c.getCLCResponse("GET", path, nil, &res)
	return res, err
}

// Get the details for a individual server.
// @serverId: name of the server being queried (e.g. WA1DTGDFEDAD0)
func (c *Client) GetServer(serverId string) (res Server, err error) {
	// Note: there exists a second way of querying a server. If @serverId is a hex UUID,
	//       then use "/v2/servers/%s/%s?uuid=True" instead.
	return c.GetServerByURI(fmt.Sprintf("/v2/servers/%s/%s", c.AccountAlias, serverId))
}

// GetServerNets returns the networks associated with the server @s.
func (c *Client) GetServerNets(s Server) (nets []Network, err error) {
	var seen = make(map[string]bool) /* map { networkId -> bool */

	// The GetNetworks() call returns only the networks visible to the current account.
	networks, err := c.GetNetworks(s.LocationId, c.AccountAlias)
	if err != nil {
		return nil, errors.Errorf("failed to query %s networks in %s: %s", c.AccountAlias, s.LocationId, err)
	}

	// In practice, the current account may be a sub-account, and the server may be
	// using a network owned by the parent's account. If that is the case,	the results will
	// be empty, and the credentials of the parent account are needed to obtain the details.
	if parentAcct := c.RegisteredAccountAlias(); parentAcct != c.AccountAlias {
		if parentNetworks, err := c.GetNetworks(s.LocationId, parentAcct); err != nil {
			return nil, errors.Errorf("failed to query %s networks in %s: %s", parentAcct, s.LocationId, err)
		} else {
			networks = append(networks, parentNetworks...)
		}
	}

	if len(networks) == 0 { // nothing found
		return nil, nil
	}

	for idx := range s.Details.IpAddresses {
		if ip := s.Details.IpAddresses[idx].Internal; ip == "" {
			/* only use the internal IPs */
		} else if net, err := NetworkByIP(ip, networks); err != nil {
			return nil, errors.Errorf("failed to identify network for %s: %s", ip, err)
		} else if net != nil && !seen[net.Id] {
			seen[net.Id] = true
			nets = append(nets, *net)
		}
	}
	return nets, nil
}

// GetIPs returns the (private, public) IP addresses associated with @serverID
func (c *Client) GetServerIPs(serverId string) (ips []string, err error) {
	srv, err := c.GetServer(serverId)
	if err != nil {
		return nil, err
	}
	return srv.IPs(), nil
}

/*
 * Create Server
 */

// CreateServerReq is the request used to create a new server instance.
type CreateServerReq struct {
	// Name of the server to create. Alphanumeric characters and dashes only.
	// Must be between 1-8 characters depending on the length of the account alias.
	// The combination of account alias and server name here must be no more than 10 characters in length.
	// (This name will be appended with a two digit number and prepended with the datacenter code
	// and account alias to make up the final server name.)
	Name string `json:"name"`

	// User-defined description of this server
	Description string `json:"description,omitempty"`

	// ID of the parent group. Retrieved from query to parent group, or by looking at the URL on the UI pages in the Control Portal.
	GroupId string `json:"groupId"`

	// ID of the server to use a source. May be the ID of a template, or when cloning, an existing server ID.
	SourceServerId string `json:"sourceServerId"`

	// Whether to create the server as managed or not. Default is false. (Ignored for bare metal servers.)
	IsManagedOs bool `json:"isManagedOS"`

	// Whether to add managed backup to the server. Must be a managed OS server. (Ignored for bare metal servers.)
	IsManagedBackup bool `json:"isManagedBackup"`

	// Primary DNS to set on the server. If not supplied the default value set on the account will be used.
	PrimaryDns string `json:"primaryDns,omitempty"`

	// Secondary DNS to set on the server. If not supplied the default value set on the account will be used.
	SecondaryDns string `json:"secondaryDns,omitempty"`

	// ID of the network to which to deploy the server. If not provided, a network will be chosen automatically.
	// If your account has not yet been assigned a network, leave this blank and one will be assigned automatically.
	NetworkId string `json:"networkId,omitempty"`

	// IP address to assign to the server. If not provided, one will be assigned automatically.
	// (Ignored for bare metal servers.)
	IpAddress string `json:"ipAddress,omitempty"`

	// Password of administrator or root user on server. If not provided, one will be generated automatically.
	Password string `json:"password,omitempty"`

	// Password of the source server, used only when creating a clone from an existing server.
	// (Ignored for bare metal servers.)
	SourceServerPassword string `json:"sourceServerPassword,omitempty"`

	// Number of processors to configure the server with (1-16) (ignored for bare metal servers)
	Cpu int `json:"cpu"`

	// ID of the vertical CPU Autoscale policy to associate the server with. (Ignored for bare metal servers.)
	CpuAutoscalePolicyId string `json:"cpuAutoscalePolicyId,omitempty"`

	// Number of GB of memory to configure the server with (1-128) (ignored for bare metal servers)
	MemoryGB int `json:"memoryGB"`

	// Whether to create a 'standard', 'hyperscale', or 'bareMetal' server
	Type string `json:"type"`

	// For standard servers, whether to use standard or premium storage.  (Ignored for bare metal servers.)
	// If not provided, will default to premium storage.
	StorageType string `json:"storageType,omitempty"`

	// ID of the Anti-Affinity policy to associate the server with. Only valid for hyperscale servers.
	AntiAffinityPolicyId string `json:"antiAffinityPolicyId,omitempty"`

	// Collection of custom field ID-value pairs to set for the server.
	CustomFields []SimpleCustomField `json:"customFields"`

	// Collection of disk parameters (ignored for bare metal servers)
	AdditionalDisks []ServerAdditionalDisk `json:"additionalDisks"`

	// Date/time that the server should be deleted (ignored for bare metal servers)
	Ttl *time.Time `json:"ttl"`

	// Collection of packages to run on the server after it has been built (ignored for bare metal servers)
	Packages []struct {
		// ID of the package to run on the server after it builds.
		PackageId string `json:"packageId"`

		// Collection of name-value pairs to specify package-specific parameters.
		Parameters struct {
			name, value string // FIXME lack of API documentation here
		} `json:"parameters"`
	} `json:"packages"`

	// Specifies the identifier for the specific configuration type of bare metal server to deploy.
	// Only required for bare metal servers. (Ignored for standard and hyperscale servers.)
	ConfigurationId string `json:"configurationId,omitempty"`

	// Specifies the OS to provision with the bare metal server.
	// Only required for bare metal servers. (Ignored for standard and hyperscale servers.)
	// Currently (Nov 3/2015), the only supported OS types are:
	// - redHat6_64Bit,
	// - centOS6_64Bit,
	// - windows2012R2Standard_64Bit,
	// - windows2012R2Datacenter_64Bit,
	// - ubuntu14_64Bit.
	OsType string `json:"osType"`
}

// Create a new server.
// @serverId: ID of the server to be deleted.
// Returns new server @url and @statusId if successful.
func (c *Client) CreateServer(req *CreateServerReq) (url, statusId string, err error) {
	var path = fmt.Sprintf("/v2/servers/%s", c.AccountAlias)

	if status, err := c.getStatusResponse("POST", path, false, req); err != nil {
		return "", "", err
	} else if link, err := extractLink(status.Links, "status"); err != nil {
		return "", "", err
	} else {
		/* Sanity checks: err != nil only if extractLink fails for expected links. */
		statusId = link.Id
		if link, err = extractLink(status.Links, "self"); err == nil {
			url = link.Href
		}
	}
	return url, statusId, nil
}

// Send the delete operation to a given server and add operation to queue.
// @serverId: ID of the server to be deleted.
func (c *Client) DeleteServer(serverId string) (statusId string, err error) {
	var path = fmt.Sprintf("/v2/servers/%s/%s", c.AccountAlias, serverId)

	return c.getStatusResponseId("DELETE", path, false, nil)
}

/*
 * OVF Import API
 */
type ImportOVF struct {
	// ID of the OVF.
	Id string

	// Name of the OVF.
	Name string

	// Number of GB of storage the server is configured with.
	StorageSizeGB int

	// Number of processors the server is configured with.
	CpuCount int

	// Number of MB of memory the server is configured with.
	MemorySizeMb int
}

// Get the list of available servers that can be imported.
// @locationId: Data center location identifier
func (c *Client) GetServerImports(locationId string) (res []ImportOVF, err error) {
	err = c.getCLCResponse("GET", fmt.Sprintf("/v2/vmImport/%s/%s/available", c.AccountAlias, locationId), nil, &res)
	return res, err
}

/*
 * Credentials
 */
type ServerCredentials struct {
	// The username of root/administrator on the server.
	// Typically "root" for Linux machines and "Administrator" for Windows.
	Username string

	// The administrator/root password used to login.
	Password string
}

// Retrieve the administrator/root password on an existing server.
// @serverId: ID of the server with the credentials to return.
func (c *Client) GetServerCredentials(serverId string) (res ServerCredentials, err error) {
	err = c.getCLCResponse("GET", fmt.Sprintf("/v2/servers/%s/%s/credentials", c.AccountAlias, serverId), nil, &res)
	return res, err
}

// Change the administrator/root password on an existing server given the current administrator/root password.
// @serverId: ID of the server to change.
// @curPass:  current password for @serverId
// @newPass:  new password for @serverId
func (c *Client) ServerChangePassword(serverId, curPass, newPass string) (statusId string, err error) {
	var op = PatchOperation{
		Op:     "set",
		Member: "password",
		Value: struct {
			// The current administrator/root password used to login.
			Current string `json:"current"`

			// The new administrator/root password to change to.
			Password string `json:"password"`
		}{curPass, newPass},
	}
	return c.patchStatus(fmt.Sprintf("/v2/servers/%s/%s", c.AccountAlias, serverId), &op)
}

// Change the number of CPU cores on an existing server.
// @serverId: ID of the server to change.
// @cpus:     number of CPUs to allocate for @serverId.
func (c *Client) ServerSetCpus(serverId, cpus string) (statusId string, err error) {
	return c.patchStatus(fmt.Sprintf("/v2/servers/%s/%s", c.AccountAlias, serverId),
		&PatchOperation{"set", "cpu", cpus})
}

// Change the amount of memory on an existing server.
// @serverId: ID of the server to change.
// @memGB:    amount of memory (in GB) to allocate.
func (c *Client) ServerSetMemory(serverId, memGB string) (statusId string, err error) {
	return c.patchStatus(fmt.Sprintf("/v2/servers/%s/%s", c.AccountAlias, serverId),
		&PatchOperation{"set", "memory", memGB})
}

// Change the description of an existing server.
// @serverId: ID of the server to change.
// @desc:     new description to use for @serverId.
func (c *Client) ServerSetDescription(serverId, desc string) error {
	return c.patch(fmt.Sprintf("/v2/servers/%s/%s", c.AccountAlias, serverId),
		&PatchOperation{"set", "description", desc})
}

// Change the description of an existing server.
// @serverId:   ID of the server to change.
// @parentUUID: UUID of new parent group for @serverId.
func (c *Client) ServerSetGroup(serverId, parentUUID string) error {
	return c.patch(fmt.Sprintf("/v2/servers/%s/%s", c.AccountAlias, serverId),
		&PatchOperation{"set", "groupId", parentUUID})
}

// SetServerDisks sets (adds, updates, removes) the disks of an existing server.
// @serverId: ID of the server to change
// @disks:    complete list of (modified) existing and (optionally) additional disks
func (c *Client) ServerSetDisks(serverId string, disks []ServerAdditionalDisk) (statusId string, err error) {
	return c.patchStatus(fmt.Sprintf("/v2/servers/%s/%s", c.AccountAlias, serverId),
		&PatchOperation{"set", "disks", disks})
}

/*
 * Server Snapshots
 */
type ServerSnapshot struct {
	// Timestamp of the snapshot (non-standard format)
	Name string

	// Collection of entity links that point to resources related to this snapshot
	Links []Link
}

// Return the single snapshot of a server, nil if none exists, or error.
// @serverId: name of the server to query
// FIXME: current (Nov 2015) CLC policy is to keep a single snapshot.
//        This may or may not change in the future.
func (c *Client) GetServerSnapshot(serverId string) (sn *ServerSnapshot, err error) {
	if server, err := c.GetServer(serverId); err != nil {
		return nil, err
	} else if len(server.Details.Snapshots) == 0 {
		return nil, nil
	} else if len(server.Details.Snapshots) > 1 {
		return nil, errors.Errorf("%s unexpectedly has more (%d) than one snapshot",
			serverId, len(server.Details.Snapshots))
	} else {
		return &server.Details.Snapshots[0], nil
	}
}

// SnapshotServer wraps CreateSnapshot, using the maximum allowed expiration period.
// If a snapshot already exists, it will be overwritten by the new one.
func (c *Client) SnapshotServer(serverId string) (statusId string, err error) {
	// CLC does not allow incremental snapshots, so delete any old ones first.
	if statusId, err := c.DeleteSnapshot(serverId); err != nil && err != ErrNoSnapshot {
		return "", err
	} else if statusId != "" {
		if status, err := c.AwaitCompletion(statusId); err != nil {
			return "", errors.Errorf("failed to query %s snapshot status: %s", serverId, err)
		} else if status != Succeeded {
			return "", errors.Errorf("failed to delete %s snapshot (status: %s)", serverId, status)
		}
	}
	return c.CreateSnapshot(serverId, 10)
}

// Send the create snapshot operation to a list of servers (along with the number of days
// to keep the snapshot for) and adds operation to queue.
// @serverId:   Server name to perform create snapshot operation on.
// @daysToKeep: Number of days to keep the snapshot(s) for (must be between 1 and 10).
func (c *Client) CreateSnapshot(serverId string, daysToKeep int) (statusId string, err error) {
	var path = fmt.Sprintf("/v2/operations/%s/servers/createSnapshot", c.AccountAlias)

	return c.getStatusResponseId("POST", path, true, &struct {
		ServerIds              []string `json:"serverIds"`
		SnapshotExpirationDays int      `json:"snapshotExpirationDays"`
	}{[]string{serverId}, daysToKeep})
}

// DeleteSnapshot deletes the server snapshot if it exists.
// @serverId: Server name to delete snapshot of.
func (c *Client) DeleteSnapshot(serverId string) (statusId string, err error) {
	var link *Link
	/*
	 * FIXME: there is no way of querying the Snapshot ID. The GetServer request
	 *        only returns the snapshot name; the ID is buried inside the URLs of
	 *        the Links array. Hence need to run 2 API requests for 1 deletion.
	 */
	if sn, err := c.GetServerSnapshot(serverId); err != nil {
		return "", err
	} else if sn == nil {
		return "", ErrNoSnapshot
	} else if link, err = extractLink(sn.Links, "delete"); err != nil {
		return "", err
	}
	return c.getStatus("DELETE", link.Href, nil)
}

// Revert server to snapshot.
// @serverId: Name of server to revert.
func (c *Client) RevertToSnapshot(serverId string) (statusId string, err error) {
	var link *Link
	/*
	 * FIXME: see above comments why this is done in this way.
	 */
	if sn, err := c.GetServerSnapshot(serverId); err != nil {
		return "", err
	} else if sn == nil {
		return "", ErrNoSnapshot
	} else if link, err = extractLink(sn.Links, "restore"); err != nil {
		return "", err
	}
	return c.getStatus("POST", link.Href, nil)
}

/*
 * Archive and restore
 */
// ArchiveServer puts @serverId into the archive
func (c *Client) ArchiveServer(serverId string) (statusId string, err error) {
	return c.serverPowerOperation("archive", serverId)
}

// RestoreServer restores @serverId into the HW Group identified by @groupId
func (c *Client) RestoreServer(serverId, groupId string) (statusId string, err error) {
	var path = fmt.Sprintf("/v2/servers/%s/%s/restore", c.AccountAlias, serverId)

	return c.getStatus("POST", path, &struct {
		TargetGroupId string `json:"targetGroupId"`
	}{groupId})
}

/*
 * Power Operations
 */
func (c *Client) serverPowerOperation(op, serverId string) (statusId string, err error) {
	var path = fmt.Sprintf("/v2/operations/%s/servers/%s", c.AccountAlias, op)

	return c.getStatusResponseId("POST", path, true, []string{serverId})
}

// Send the pause operation to a server and add operation to queue.
// @serverId: Name of server to pause.
func (c *Client) PauseServer(serverId string) (statusId string, err error) {
	return c.serverPowerOperation("pause", serverId)
}

// Send the power-on operation to a server and add operation to queue.
// @serverId: Name of server to power on.
func (c *Client) PowerOnServer(serverId string) (statusId string, err error) {
	return c.serverPowerOperation("powerOn", serverId)
}

// Send the (hard) power-off operation to a server and add operation to queue.
// @serverId: Name of server to power off.
func (c *Client) PowerOffServer(serverId string) (statusId string, err error) {
	return c.serverPowerOperation("powerOff", serverId)
}

// Send the (soft) shut-down operation to a server and add operation to queue.
// @serverId: Name of server to shut down.
func (c *Client) ShutdownServer(serverId string) (statusId string, err error) {
	return c.serverPowerOperation("shutDown", serverId)
}

// Send the reboot operation to a server and add operation to queue.
// @serverId: Name of server to reboot.
func (c *Client) RebootServer(serverId string) (statusId string, err error) {
	return c.serverPowerOperation("reboot", serverId)
}

// Send the reset operation to a server and add operation to queue.
// @serverId: Name of server to reset.
func (c *Client) ResetServer(serverId string) (statusId string, err error) {
	return c.serverPowerOperation("reset", serverId)
}

// Send the start-maintenance operation to a server and add operation to queue.
// @serverId: Name of server to change.
func (c *Client) ServerStartMaintenance(serverId string) (statusId string, err error) {
	return c.serverPowerOperation("startMaintenance", serverId)
}

// Send the stop-maintenance operation to a server and add operation to queue.
// @serverId: Name of server to change.
func (c *Client) ServerStopMaintenance(serverId string) (statusId string, err error) {
	return c.serverPowerOperation("stopMaintenance", serverId)
}

type MaintenanceMode struct {
	// ID of server to set maintenance mode on or off
	Id string `json:"id"`

	// Indicator of whether to place server in maintenance mode or not
	InMaintenanceMode bool `json:"inMaintenanceMode"`
}

// Send a specified setting for maintenance mode to a server and add operation to queue.
// @serverId: Name of server to change.
// @enable:   Whether to enable (true) or disable (false) Maintenance Mode on @serverId.
func (c *Client) ServerSetMaintenance(serverId string, enable bool) (statusId string, err error) {
	var path = fmt.Sprintf("/v2/operations/%s/servers/setMaintenance", c.AccountAlias)

	return c.getStatusResponseId("POST", path, true, &struct {
		Servers []MaintenanceMode `json:"servers"`
	}{[]MaintenanceMode{{serverId, enable}}})
}

/*
 * Adding/removing secondary network adapters.
 * FIXME: the status response returns a different object than the regular queue status
 *        operation, which requires patching up here (experimental API).
 */
const (
	/* Poll interval in seconds for adding/removing a secondary network adapter */
	change_nic_poll = 1 * time.Second

	/* Maximum acceptable time in seconds to poll for secondary NIC change in minutes */
	change_nic_timeout = 3 * time.Minute
)

type ChangeNicResponse struct {
	// GUID for the item in the queue for completion
	OperationId string

	// Link to review status of the operation,
	Uri string
}

// The object returned from a GET at ChangeNicResponse.Uri
type ChangeNicStatus struct {
	// This seems to be "secondaryNetworkAdapter"
	RequestType string

	// This starts with "queued" and reaches "succeeded" when done
	Status QueueStatus

	// Use maps for the rest: there is no documentation currently
	Summary, Source map[string]string
}

// Helper function to poll the queue used for adding/removing secondary network interfaces.
// Since this uses a diffrent API, the standard Queue -> Get Status can not be used.
func (c *Client) changeNic(verb, path string, reqModel interface{}) (err error) {
	var res ChangeNicResponse
	var s ChangeNicStatus

	if err = c.getCLCResponse(verb, path, reqModel, &res); err == nil {
		for start := time.Now(); s.Status != Succeeded; time.Sleep(change_nic_poll) {
			if err = c.getCLCResponse("GET", res.Uri, nil, &s); err != nil {
				break
			} else if s.Status == Failed {
				return errors.Errorf("request %s %s failed", verb, path)
			} else if time.Since(start) > change_nic_timeout {
				return errors.Errorf("request %s %s timed out after %s", verb,
					path, time.Since(start))
			}
		}
	}
	return err
}

// Add secondary network adapter to server.
// @serverId: ID of the server to change
// @netId:    (Hex) ID of the network to connect to (must be different from server's existing ones)
// @ip:       Optional IP address to claim on the network @netId
func (c *Client) ServerAddNic(serverId, netId, ip string) (err error) {
	return c.changeNic("POST", fmt.Sprintf("/v2/servers/%s/%s/networks", c.AccountAlias, serverId), struct {
		// (Hex) ID of the network.
		NetworkId string `json:"networkId"`

		// Optional IP address for the networkId
		IpAddress string `json:"ipAddress,omitempty"`
	}{netId, ip})
}

// Remove secondary network adapter from server.
// @serverId: ID of the server to change
// @netId:    ID of the network
func (c *Client) ServerDelNic(serverId, netId string) (err error) {
	return c.changeNic("DELETE", fmt.Sprintf("/v2/servers/%s/%s/networks/%s", c.AccountAlias, serverId, netId), nil)
}
