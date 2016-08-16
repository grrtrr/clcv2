package clcv2

import "fmt"

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

// SBSgetOsTypes returns the list of Operating System Types supported by the Simple Backup Service.
func (c *Client) SBSgetOsTypes() (res []string, err error) {
	err = c.getSBSResponse("GET", "osTypes", nil, &res)
	return
}

// SBSpolicy describes a single SBS policy.
type SBSpolicy struct {
	// The total number of Policies associated with the Account
	TotalCount int

	// The maximum number of results requested
	Limit int

	// The next position in the list of results for a subsequent call
	NextOffset int

	// The starting position in the list of results
	Offset int

	// An array of the Policies associated with the Account
	Results []SBSAccountPolicy
}

// SBSAccountPolicy contains the actual SBS account policy information.
type SBSAccountPolicy struct {
	// The name of the Policy
	Name string

	// The account alias that the Policy belongs to
	ClcAccountAlias string

	// The unique Id associated with this Policy
	PolicyID string

	// The OS Type - 'Linux' or 'Windows'
	OsType string

	// The backup frequency of the Policy (= duration between backups)
	BackupIntervalHours int

	// A list of the directories that the Policy includes in each backup
	Paths []string

	// A list of the directories that the Policy excludes from backup
	ExcludedDirectoryPaths []string

	// The number of days backup data will be retained
	RetentionDays int

	// The status of the backup Policy. Either 'ACTIVE' or 'INACTIVE'.
	Status string
}

// SBSgetPolicy returns the single Policy associated with @policyID, or an error.
func (c *Client) SBSgetPolicy(policyID string) (res SBSAccountPolicy, err error) {
	err = c.getSBSResponse("GET", fmt.Sprintf("accountPolicies/%s", policyID), nil, &res)
	return
}

// SBSgetPolicies returns the list of SBS backup policies associated with an account.
// NOTE: currently ignoring the query parameters, which seem to be more useful for JavaScript rendering.
func (c *Client) SBSgetPolicies() (res SBSpolicy, err error) {
	err = c.getSBSResponse("GET", "accountPolicies", nil, &res)
	return
}
