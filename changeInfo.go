package clcv2

import "time"

type ChangeInfo struct {
	// Date/time of resource creation
	CreatedDate time.Time

	// Who created the resource
	CreatedBy string

	// Date/time the resource was last updated
	ModifiedDate time.Time

	// Who modified the resource last
	ModifiedBy string
}
