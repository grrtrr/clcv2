/*
 * API v2.0 Links Framework
 */
package clcv2

import (
	"strings"
	"fmt"
)

// Link adds hyperlink information to resources.
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

// Extract Links whose 'Rel' field matches @rel_type.
func ExtractLinks(from []Link, rel_type string) (res []Link) {
	for _, l := range from {
		if l.Rel == rel_type {
			res = append(res, l)
		}
	}
	return
}

// Extract first Link whose 'Rel' field matches @rel_type, return nil if none found.
func extractLink(from []Link, rel_type string) (l *Link, err error) {
	if links := ExtractLinks(from, rel_type); len(links) > 0 {
		// FIXME: maybe warn here if there is more than 1 match
		l = &links[0]
	} else {
		err = fmt.Errorf("No link with Rel=%s found in %+v", rel_type, from)
	}
	return
}
