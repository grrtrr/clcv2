package clcv2

import (
_	"time"
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

// Get the details for a individual server group and any sub-groups and servers that it contains.
// @groupId: ID of the group being queried.
func (c *Client) GetGroup(groupId string) (res Group, err error) {
	path := fmt.Sprintf("/v2/groups/%s/%s", c.AccountAlias, groupId)
	err = c.getResponse("GET", path, nil, &res)
	return
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
