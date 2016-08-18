package clcv2

import (
	"fmt"
	"net/url"
	"time"

	"github.com/kr/pretty"
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
type SBSAccountPolicy struct {
	// The backup frequency of the Policy (= duration between backups)
	BackupIntervalHours int

	// The account alias that the Policy belongs to
	ClcAccountAlias string

	// A list of the directories that the Policy excludes from backup
	ExcludedDirectoryPaths []string

	// The name of the Policy
	Name string

	// The OS Type - 'Linux' or 'Windows'
	OsType string

	// A list of the directories that the Policy includes in each backup
	Paths []string

	// The unique Id associated with this Policy
	PolicyID string

	// The number of days backup data will be retained
	RetentionDays int

	// The status of the backup Policy. Either 'ACTIVE' or 'INACTIVE'.
	Status string
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

// SBSgetPolicy returns the single Policy associated with @policyID, or an error.
func (c *Client) SBSgetPolicy(policyID string) (res SBSAccountPolicy, err error) {
	err = c.getSBSResponse("GET", fmt.Sprintf("accountPolicies/%s", policyID), nil, &res)
	return
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
	return nil, fmt.Errorf("no Server Policy %q found for this account", serverPolicyId)
}

// SBSpatchServerPolicyStatus sets the status of the specified Server Policy to @newValue.
func (c *Client) SBSpatchServerPolicyStatus(srvPolicyID, newValue string) (p *SBSServerPolicy, err error) {
	p, err = c.SBSgetServerPolicy(srvPolicyID)
	if err != nil {
		return
	}
	pretty.Println(*p)

	err = c.getCLCResponse("PATCH",
		fmt.Sprintf("accountPolicies/%s/serverPolicies/%s", p.AccountPolicyID, p.ID),
		&struct {
			Op    string `json:"op"`
			Path  string `json:"path"`
			Value string `json:"value"`
		}{"add", "/status", newValue}, p)
	if p != nil {
		pretty.Println(*p)
	}
	return
}

// SBSRestorePoint captures the details of a single restore point
type SBSRestorePoint struct {
	// Unique restore point identifier
	RestorePointID string

	// Unique policy identifier
	PolicyId string

	// Days of retention applied to the restore point
	RetentionDays int

	// Timestamp of backup completion
	BackupFinishedDate string

	// Timestamp or retention expiration
	RetentionExpiredDate string

	// 'SUCCESS', 'PARTIAL_SUCCESS', 'FAILED', or 'CANCELLED'
	RestorePointCreationStatus string

	// Number of backup files transferred to storage
	FilesTransferredToStorage int

	// Total bytes of backup data sent to storage
	BytesTransferredToStorage int

	// Number of backup files that failed transfer to storage
	FilesFailedTransferToStorage int

	// Total bytes of backup data that failed transfer to storage
	BytesFailedToTransfer int

	// Number of unchanged files not requiring retransfer to storage
	UnchangedFilesNotTransferred int

	// Total bytes of unchanged data not requiring retransfer to storage
	UnchangedBytesInStorage int

	// Number of files removed from local disk
	FilesRemovedFromDisk int

	// Total bytes of data removed from local disk
	BytesInStorageForItemsRemoved int

	// Number of files currently in storage for the restore point
	NumberOfProtectedFiles int

	// Timestamp of backup start
	BackupStartedDate string
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
