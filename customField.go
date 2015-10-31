package clcv2

import "fmt"

/* Custom field as it appears embedded in other structures. */
type CustomField struct {
	// Unique ID of the custom field
	Id           string

	// Friendly name of the custom field
	Name         string

	// Underlying value of the field
	Value        string

	// Shown value of the field
	DisplayValue string
}

/* Custom fields as associated with the account */
type AccountCustomField struct {
	// Unique identifier of the custom field.
	Id		string

	// Friendly name of the custom field as it appears in the UI.
	Name		string

	// Boolean value representing whether or not a value is required for this custom field.
	IsRequired	bool

	// The type of custom field defined. Will be either
	// - "text"     (free-form text field),
	// - "checkbox" (boolean value), or
	// - "option"   (drop-down list).
	Type		string

	// Array of name-value pairs corresponding to the options defined for this field.
	// (Empty for "text" or "checkbox" field types.)
	Options		[]struct { Name, Value string }
}


// Retrieve the custom field(s) defined for a given account.
// @acctAlias: Short code for a particular account (leave empty to use default account alias)
func (c *Client) GetCustomFields(acctAlias string) (res []AccountCustomField, err error) {
	if acctAlias == "" {
		acctAlias = c.AccountAlias
	}
	err = c.getResponse("GET", fmt.Sprintf("/v2/accounts/%s/customFields", acctAlias), nil, &res)
	return
}
