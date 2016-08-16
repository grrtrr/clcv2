package clcv2

// Patch operation describes a single PATCH operation to be performed on a CLCv2 resource.
type PatchOperation struct {
	// The operation to perform on a given property of the resource.
	Op string `json:"op"`

	// The property of the resource to perform the operation on.
	Member string `json:"member"`

	// The value to patch - depends on the type of operation.
	Value interface{} `json:"value"`
}

// Run patch operation(s) @ops and return Link status.
func (c *Client) patchStatus(path string, ops ...*PatchOperation) (statusId string, err error) {
	return c.getStatus("PATCH", path, ops)
}

// Like patchStatus(), but without statusId. For those patch operations that return '204 No Content'.
func (c *Client) patch(path string, ops ...*PatchOperation) error {
	return c.getCLCResponse("PATCH", path, ops, nil)
}
