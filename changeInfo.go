package clcv2

import "time"

// ChangeInfo records resource creation and modification details.
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
