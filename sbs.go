package clcv2

import (
	"fmt"
	"net/url"
	"strconv"
	"time"
)

/*
 * Simple Backup API
 */
const (
	// SBS root URL
	SBSurl = "https://api.backup.ctl.io/clc-backup-api/api/"
)

// getCLCResponse performs a CLC v2 main API request
// @verb: Http verb to use
// @path: relative to %SBSurl
func (c *Client) getSBSResponse(verb, path string, reqModel, resModel interface{}) (err error) {
	return c.getResponse(SBSurl+path, verb, reqModel, resModel)
}

// SBSregion represents SBS storage region information
type SBSregion struct {
	// The name of the Storage Region
	Name string

	// The label associated with the Storage Region
	RegionLabel string

	// A list of messages associated with the Storage Region
	Messages []string
}

// SBSgetAllRegions retrieves a list of backup storage regions which are available in Simple Backup Service.
func (c *Client) SBSgetAllRegions() (res []SBSregion, err error) {
	err = c.getSBSResponse("GET", "regions", nil, &res)
	return
}

// SBSgetDatacenters returns a list of CLC data centres.
func (c *Client) SBSgetDatacenters() (res []string, err error) {
	err = c.getSBSResponse("GET", "datacenters", nil, &res)
	return
}

// SBSgetServersByDatacenter returns a list of servers associated with data center @dc.
func (c *Client) SBSgetServersByDatacenter(dc string) (res []string, err error) {
	var u = url.URL{Path: fmt.Sprintf("datacenters/%s/servers", dc)}
	err = c.getSBSResponse("GET", u.EscapedPath(), nil, &res)
	return
}

// SBSgetOsTypes returns the list of Operating System Types supported by the Simple Backup Service.
func (c *Client) SBSgetOsTypes() (res []string, err error) {
	err = c.getSBSResponse("GET", "osTypes", nil, &res)
	return
}

// SBSAccountPolicy contains the actual SBS account policy information.
// It is used to list existing Account Policies, and to create new ones (hence the json tags).
type SBSAccountPolicy struct {
	// The backup frequency of the Policy (= duration between backups)
	BackupIntervalHours int `json:"backupIntervalHours"`

	// The account alias that the Policy belongs to
	ClcAccountAlias string `json:"clcAccountAlias"`

	// A list of the directories that the Policy excludes from backup
	ExcludedDirectoryPaths []string `json:"excludedDirectoryPaths"`

	// The name of the Policy
	Name string `json:"name"`

	// The OS Type - 'Linux' or 'Windows'
	OsType string `json:"osType"`

	// A list of the directories that the Policy includes in each backup
	Paths []string `json:"paths"`

	// The unique Id associated with this Policy
	PolicyID string `json:"policyId,omitempty"`

	// The number of days backup data will be retained
	RetentionDays int `json:"retentionDays"`

	// The status of the backup Policy. Either 'ACTIVE' or 'INACTIVE'.
	Status string `json:"status,omitempty"`
}

// SBScreatePolicy creates a new Account Policy
func (c *Client) SBScreatePolicy(req *SBSAccountPolicy) (res SBSAccountPolicy, err error) {
	err = c.getSBSResponse("POST", "accountPolicies", req, &res)
	return
}

// SBSupdatePolicy updates an existing Account Policy
func (c *Client) SBSupdatePolicy(policyID string, req *SBSAccountPolicy) (res SBSAccountPolicy, err error) {
	err = c.getSBSResponse("PUT", fmt.Sprintf("accountPolicies/%s", policyID), req, &res)
	return
}

// SBSgetPolicy returns the single Policy associated with @policyID, or an error.
func (c *Client) SBSgetPolicy(policyID string) (res SBSAccountPolicy, err error) {
	err = c.getSBSResponse("GET", fmt.Sprintf("accountPolicies/%s", policyID), nil, &res)
	return
}

// SBSgetPolicies returns the list of SBS backup policies associated with an account.
func (c *Client) SBSgetPolicies() ([]SBSAccountPolicy, error) {
	// Note: we do not paging for this API, so just wrap it in anonymous struct.
	var result struct {
		Results []SBSAccountPolicy
	}
	err := c.getSBSResponse("GET", "accountPolicies", nil, &result)
	return result.Results, err
}

// SBSgetEligiblePolicies returns the list of Account Policies eligible for the specified @server.
func (c *Client) SBSgetEligiblePolicies(server string) ([]SBSAccountPolicy, error) {
	var result struct {
		Results []SBSAccountPolicy
	}
	err := c.getSBSResponse("GET", fmt.Sprintf("accountPolicies/servers/%s", server), nil, &result)
	return result.Results, err
}

// SBSSingleServerPolicy represents the policy associated with a server
type SBSSingleServerPolicy struct {
	// The name of the Policy
	Name string

	// The account alias that the Policy belongs to
	ClcAccountAlias string

	// Unique ID of an account policy
	AccountPolicyId string

	// Unique server policy identifier
	ServerPolicyID string

	// Unique server identifier
	ServerID string

	// OS Type - 'Linux' or 'Windows'
	OsType string

	// The status of the backup Policy. Either 'ACTIVE' or 'INACTIVE'
	AccountPolicyStatus string

	// Status of the backup policy. 'ACTIVE','INACTIVE','PROVISIONING','ERROR','DELETED'
	ServerPolicyStatus []string

	// Indicates if the account policy or server policy are active/inactive
	EligibleForBackup bool

	// A list of the directories that the Policy includes in each backup
	Paths []string

	// The number of days backup data will be retained
	RetentionDays int

	// The backup frequency of the Policy specified in hours
	BackupIntervalHours int

	// Region where backups are stored. "US EAST", "US WEST", "CANADA", "GREAT BRITAIN", "GERMANY", "APAC"
	StorageRegion string

	// Not currently used
	BackupProvider string
}

// SBSServerPolicy represents a single SBS server policy, associated to an Account Policy
type SBSServerPolicy struct {
	// The Server Policy ID
	ID string `json:"serverPolicyId"`

	// Unique Id of the Account Policy
	AccountPolicyID string

	// Unique server name
	ServerID string

	// Region where backups are stored
	StorageRegion string

	// The account alias that the Policy belongs to
	ClcAccountAlias string

	// The status of the backup Policy. 'ACTIVE', 'INACTIVE', 'PROVISIONING', 'ERROR', 'DELETED'
	Status string

	// Date all data retention will elapse; unsubscribedDate+retentionDays
	ExpirationDate int

	// Date policy was inactivated
	UnsubscribedDate int

	// Not currently used
	storageAccountID string
}

// SBScreateServerPolicy creates a new Server Policy for the given @server, Account Policy ID, and @region.
func (c *Client) SBScreateServerPolicy(acPolicyID, server, region string) (res SBSServerPolicy, err error) {
	err = c.getSBSResponse("POST", fmt.Sprintf("accountPolicies/%s/serverPolicies", acPolicyID),
		struct {
			Account string `json:"clcAccountAlias"`
			Server  string `json:"serverId"`
			Region  string `json:"storageRegion"`
		}{c.AccountAlias, server, region}, &res)
	return
}

// SBSdeleteServerPolicy deletes the Server Policy specified by @srvPolicyID
func (c *Client) SBSdeleteServerPolicy(srvPolicyID string) error {
	p, err := c.SBSgetServerPolicy(srvPolicyID)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("accountPolicies/%s/serverPolicies/%s", p.AccountPolicyID, p.ID)
	return c.getSBSResponse("DELETE", path, nil, nil)
}

// SBSgetServerPolicies returns a list of Server Policies associated to an Account Policy
func (c *Client) SBSgetServerPolicies(acPolicyId string) ([]SBSServerPolicy, error) {
	var result struct {
		Results []SBSServerPolicy
	}
	err := c.getSBSResponse("GET", fmt.Sprintf("accountPolicies/%s/serverPolicies", acPolicyId), nil, &result)
	return result.Results, err
}

// SBSgetServerPolicyDetails returns SBS policy details associated with a single @server.
func (c *Client) SBSgetServerPolicyDetails(server string) (res []SBSServerPolicy, err error) {
	err = c.getSBSResponse("GET", fmt.Sprintf("serverPolicyDetails?serverId=%s", server), nil, &res)
	return
}

// SBSgetServerPolicy list SBS server policy details of the given @serverPolicyId
func (c *Client) SBSgetServerPolicy(serverPolicyId string) (*SBSServerPolicy, error) {
	acPolicies, err := c.SBSgetPolicies()
	if err != nil {
		return nil, err
	}
	for _, acPolicy := range acPolicies {
		srvPolicies, err := c.SBSgetServerPolicies(acPolicy.PolicyID)
		if err != nil {
			return nil, err
		}
		for _, srvPolicy := range srvPolicies {
			if srvPolicy.ID == serverPolicyId {
				return &srvPolicy, nil
			}
		}
	}
	return nil, fmt.Errorf("Server Policy %q not found for %s account", serverPolicyId, c.AccountAlias)
}

// SBSpatchServerPolicyStatus sets the status of the specified Server Policy to @newValue.
// Note: this command seems unsupported. In my tests, it always returned 200 and did nothing.
func (c *Client) SBSpatchServerPolicyStatus(srvPolicyID, newValue string) (p *SBSServerPolicy, err error) {
	p, err = c.SBSgetServerPolicy(srvPolicyID)
	if err != nil {
		return
	}

	err = c.getSBSResponse("PATCH",
		fmt.Sprintf("accountPolicies/%s/serverPolicies/%s", p.AccountPolicyID, p.ID),
		struct {
			// According to v2 documentation 2016-08-22, only supported op is 'replace',
			// and only supported path is '/status'
			Op    string `json:"op"`
			Path  string `json:"path"`
			Value string `json:"value"`
		}{"replace", "/status", newValue}, p)
	return
}

// SBSRestorePoint captures the details of a single restore point
type SBSRestorePoint struct {
	// Unique restore point identifier
	RestorePointID string

	// The Server Policy ID associated with this Restore Point
	PolicyId string

	// Days of retention applied to the restore point
	RetentionDays int

	// Timestamp of backup completion
	BackupFinishedDate time.Time

	// Timestamp or retention expiration
	RetentionExpiredDate time.Time

	// 'SUCCESS', 'PARTIAL_SUCCESS', 'FAILED', or 'CANCELLED'
	RestorePointCreationStatus string

	// Number of backup files transferred to storage
	FilesTransferredToStorage uint64

	// Total bytes of backup data sent to storage
	BytesTransferredToStorage uint64

	// Number of backup files that failed transfer to storage
	FilesFailedTransferToStorage uint64

	// Total bytes of backup data that failed transfer to storage
	BytesFailedToTransfer uint64

	// Number of unchanged files not requiring retransfer to storage
	UnchangedFilesNotTransferred uint64

	// Total bytes of unchanged data not requiring retransfer to storage
	UnchangedBytesInStorage uint64

	// Number of files removed from local disk
	FilesRemovedFromDisk uint64

	// Total bytes of data removed from local disk
	BytesInStorageForItemsRemoved uint64

	// Number of files currently in storage for the restore point
	NumberOfProtectedFiles uint64

	// Timestamp of backup start
	BackupStartedDate time.Time
}

// SBSgetServerPolicyDetails returns SBS restore point details for a given Account and Server Policy.
// @acPolicy:  account policy ID
// @srvPolicy: server policy ID (it derives from @acPolicy, hence I don't understand why @acPolicy is used)
// @start:     start time (date) of the backup to list
// @end:       end time (date) of the backup to list
func (c *Client) SBSgetRestorePointDetails(acPolicy, srvPolicy string, start, end time.Time) ([]SBSRestorePoint, error) {
	var path = fmt.Sprintf("accountPolicies/%s/serverPolicies/%s/restorePointDetails?"+
		"backupFinishedStartDate=%s&backupFinishedEndDate=%s",
		acPolicy, srvPolicy, start.Format("2006-01-02"), end.Format("2006-01-02"))
	var result struct {
		Results []SBSRestorePoint
	}

	err := c.getSBSResponse("GET", path, nil, &result)
	return result.Results, err
}

// SBSgetServerStorageUsage returns the number of bytes used by the specified Server Policy on a given @day.
func (c *Client) SBSgetServerStorageUsage(acPolicy, srvPolicy string, day time.Time) (uint64, error) {
	var path = fmt.Sprintf("accountPolicies/%s/serverPolicies/%s/storedData?searchDate=%s",
		acPolicy, srvPolicy, day.Format("2006-01-02"))
	var result struct {
		GigaBytesStored string // I don't understand why they are doing this.
		BytesStored     string // Why are they converting numeric quantities into strings?
	}

	if err := c.getSBSResponse("GET", path, nil, &result); err != nil {
		return 0, err
	}
	val, err := strconv.ParseUint(result.BytesStored, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid result %s GB / %sB: %s", result.GigaBytesStored, result.BytesStored, err)
	}
	return val, err
}
