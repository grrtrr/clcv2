/*
 * API v2.0 Links Framework
 */
package clcv2

import (
	"strings"
	"fmt"
)

type Link struct {
	// The link type (depends on context)
	Rel	string

	// Address of the resource.
	Href	string

	/*
         * Optional Fields
         */
	// Unique ID of the resource.
	Id	string

	// Friendly name of the resource.
	Name	string

	// Valid HTTP verbs that can act on this resource.
	// If none are explicitly listed, GET is assumed to be the only one.
	Verbs	[]string

}

func (l *Link) String() string {
	return fmt.Sprintf("%s: %s %s", l.Rel, l.Href, strings.Join(l.Verbs, ", "))
}
