package clcv2

import (
	"time"
	"fmt"
)

type Group struct {
	// ID of the group being queried
	Id		string

	// User-defined name of the group
	Name		string

	// User-defined description of this group
	Description	string

	// Data center location identifier
	LocationId	string

	//Group type which could include system types like "archive"
	Type		string

	// Describes if group is online or not
	Status		string

	// Number of servers this group contains
	Serverscount	int

	// Refers to this same entity type for each sub-group
	Groups		[]Group

	// Collection of entity links that point to resources related to this group
	Links		[]Link

	// Describes "created" and "modified" details
	ChangeInfo	ChangeInfo

	// complexDetails about any custom fields and their values
	CustomFields	[]CustomField
}

// Get the details of an individual server and any sub-groups and servers that it contains.
// @groupId: ID of the group being queried.
func (c *Client) GetGroup(groupId string) (res Group, err error) {
	path := fmt.Sprintf("/v2/groups/%s/%s", c.AccountAlias, groupId)
	err = c.getResponse("GET", path, nil, &res)
	return
}

// Gets a list of all groups with the specified search criteria.
// @location:  The data center location to query for groups.
func (c *Client) GetGroups(location string) (rootNode Group, err error) {
	dc, err := c.GetDatacenter(location, true)
	if err != nil {
		return
	}

	gl, err := extractLink(dc.Links, "group")
	if err != nil {
		return
	}
	return c.GetGroup(gl.Id)
}

// Do a depth-first traversal of the tree to find a specific node.
// @root:  where to start the search at
// @found: function indicating whether the passed Group is the one looked for
func FindGroupNode(root *Group, found func(*Group) bool) *Group {
	if found(root) {
		return root
	}
	for _, c := range root.Groups {
		if node := FindGroupNode(&c, found); node != nil {
			return node
		}
	}
	return nil
}

// Helper function to recursively add @g to @res if @found(g) returns true.
func visitGroup(g *Group, res chan *Group, found func(*Group) bool) {
	if found(g) {
		res <- g
	}
	for i := range g.Groups {
		visitGroup(&g.Groups[i], res, found)
	}
}

// Look for all Hardware Groups satisfying a given criterion
// @location:  The data center location to query for groups.
// @found:     Function returning true if the passed Group qualifies.
// Returns array of pointers to Group; or error on failure.
func (c *Client) GetGroupsFiltered(location string, found func(*Group) bool) (res []*Group, err error) {
	var rootNode Group

	if rootNode, err = c.GetGroups(location); err == nil {
		resultChan := make(chan *Group)
		go func() {
			visitGroup(&rootNode, resultChan, found)
			close(resultChan)
		}()
		for result := range resultChan {
			res = append(res, result)
		}
	}
	return
}

// Look for a (the first) Hardware Group satisfying a given criterion
// @location:  The data center location to query for groups.
// @found:     Function returning true if the passed Hardware Group is the one looked for.
// Returns pointer to Group, nil if not found; or error on failure.
func (c *Client) GetGroupFiltered(location string, found func(*Group) bool) (res *Group, err error) {
	if groups, err := c.GetGroupsFiltered(location, found); err != nil {
		return nil, err
	} else if len(groups) == 1 {
		res = groups[0]
	} else if len(groups) > 1 {
		return nil, fmt.Errorf("ambiguous - %d matching groups found at %s", len(groups), location)
	}
	return
}

// Look up Hardware Group by @name and @location
func (c *Client) GetGroupByName(name, location string) (*Group, error) {
	return c.GetGroupFiltered(location, func(g *Group) bool { return g.Name == name })
}

// Look up Hardware Group by @uuid and @location
// The @location is required, since there is no global 'resolveGroup(uuid)' function.
func (c *Client) GetGroupByUUID(uuid, location string) (*Group, error) {
	return c.GetGroupFiltered(location, func(g *Group) bool { return g.Id == uuid })
}

// Create a new Hardware Group.
// @name:   Name of the group to create.
// @parent: The unique identifier of the parent group.
// @desc:   User-defined description of this group.
// @cf:     Optional array of Custom Fields to set.
func (c *Client) CreateGroup(name, parent, desc string, cf []SimpleCustomField) (res Group, err error) {
	req := struct {
		Name		string			`json:"name"`
		Description	string			`json:"description"`
		ParentGroupId	string			`json:"parentGroupId"`
		customFields	[]SimpleCustomField	`json:"customFields"`
	} { name, desc, parent, cf }
	err = c.getResponse("POST", fmt.Sprintf("/v2/groups/%s", c.AccountAlias), &req, &res)
	return
}

// Change the name of an existing group.
// @groupId: ID of the group to update
// @newName: new name for @groupId.
func (c *Client) GroupSetName(groupId, newName string) error {
	return c.patch(fmt.Sprintf("/v2/groups/%s/%s", c.AccountAlias, groupId),
		       &PatchOperation{ "set", "name", newName })
}

// Change the description of an existing group.
// @groupId: ID of the group to update
// @newDesc: new description for @groupId.
func (c *Client) GroupSetDescription(groupId, newDesc string) error {
	return c.patch(fmt.Sprintf("/v2/groups/%s/%s", c.AccountAlias, groupId),
		       &PatchOperation{ "set", "description", newDesc })
}

// Change the parent HW group of an existing group.
// @groupId: ID of the group to update
// @parentUUID: UUID of new parent group for @groupId.
func (c *Client) GroupSetParent(groupId, parentUUID string) error {
	return c.patch(fmt.Sprintf("/v2/groups/%s/%s", c.AccountAlias, groupId),
		       &PatchOperation{ "set", "parentGroupId", parentUUID })
}


// Send the delete operation to a given group and adds operation to queue.
// This operation will delete the group and all servers and groups underneath it.
// @groupId: UUID of the group to be deleted.
func (c *Client) DeleteGroup(groupId string) (statusId string, err error) {
	return c.getStatus("DELETE", fmt.Sprintf("/v2/groups/%s/%s", c.AccountAlias, groupId), nil)
}

// Print group hierarchy starting at @g, using initial indentation @indent.
func PrintGroupHierarchy(g *Group, indent string) {
	var groupLine string

	if g.Type == "default" {
		groupLine = fmt.Sprintf("%s%s/", indent, g.Name)
	} else {
		groupLine = fmt.Sprintf("%s[%s]/", indent, g.Name)
	}
	fmt.Printf("%-70s %s\n", groupLine, g.Id)

	for _, l := range g.Links {
		if l.Rel == "server" {
			fmt.Printf("%s", indent + "    ")
			fmt.Printf("%s\n", l.Id)
		}
	}

	for idx := range g.Groups {
		PrintGroupHierarchy(&g.Groups[idx], indent + "    ")
	}
}

/*
 * Group Billing Details
 */
type GroupBillingDetails struct {
	// Date that this billing information applies to
	Date	time.Time
	// List of groups (with the first being the queried group) in the requested hierarchy.
	// The keys of the map are the group Ids.
	Groups	map[string]struct{
		// User-defined name of the group
		Name	string

		// Collection of servers in this group, each with billing information.
		// The keys of this map are the server names.
		Servers	map[string]ServerBillingDetails
	}
}

// Get the current and estimated charges for each server in a designated group hierarchy.
// @groupId: ID of the group being queried.
func (c *Client) GetGroupBillingDetails(groupId string) (res GroupBillingDetails, err error) {
	path := fmt.Sprintf("/v2/groups/%s/%s/billing", c.AccountAlias, groupId)
	err = c.getResponse("GET", path, nil, &res)
	return
}

/*
 * Group Scheduled Activities
 */
type GroupScheduledActivity struct {
	// ID of the group
	Id			string

	// Data center location identifier
	LocationId		string

	// Change history
	ChangeInfo		ChangeInfo

	// Collection of entity links that point to resources related to this data center
	Links			[]Link

	// State of scheduled activity: on or off
	Status			string

	// Type of activity: archive, createsnapshot, delete, deletesnapshot, pause, poweron, reboot, shutdown
	Type			string

	// Time when scheduled activity should start
	BeginDateUtc		time.Time

	// How often to repeat: never, daily, weekly, monthly, customWeekly
	Repeat			string

	// An array of strings for the days of the week: sun, mon, tue, wed, thu, fri, sat
	CustomWeeklyDays	[]string

	// When the scheduled activities are set to expire: never, afterDate, afterCount
	Expire			string

	// Number of times scheduled activity should run before expiring
	ExpireCount		int

	// When the scheduled activity should expire
	ExpireDateUtc		time.Time

	// To display in local time
	TimeZoneOffset		string

	// True: scheduled activity has expired. False: scheduled activity is active
	IsExpired		bool

	// Last time scheduled activity was run
	LastOccurrenceDateUtc	time.Time

	// How many times scheduled activity has been run
	OccurrenceCount		int

	// When the next scheduled activty will be run
	NextOccurrenceDateUtc	time.Time
}

// Get the scheduled activities associated with a group.
// @groupId: ID of the group being queried.
func (c *Client) GetGroupScheduledActivities(groupId string) (res []GroupScheduledActivity, err error) {
	path := fmt.Sprintf("/v2/groups/%s/%s/ScheduledActivities", c.AccountAlias, groupId)
	err = c.getResponse("GET", path, nil, &res)
	return
}
