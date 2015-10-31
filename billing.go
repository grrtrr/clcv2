package clcv2

import (
	"time"
	"fmt"
)

type InvoiceData struct {
	// ID of the account alias being queried
	Id			string

	// payment terms associated with the account
	Terms			string

	// Description of the account name
	CompanyName		string

	// Short code for a particular account
	AccountAlias		string

	// Short code for a particular account that receives the bill for the accountAlias usage
	PricingAccountAlias	string

	// Short code for the parent account alias associated with the account alias being queried
	ParentAccountAlias	string

	// First line of the address associated with accountAlias
	Address1		string

	// Second line of the address associated with accountAlias
	// FIXME: not found in v2 output (Oct 2015)
	Address2		string

	// City associated with the accountAlias
	City			string

	// State or province associated with the accountAlias
	StateProvince		string

	// Postal code associated with the accountAlias
	PostalCode		string

	// Billing email address associated with the accountAlias
	BillingContactEmail	string

	// Additional billing email address associated with the accountAlias
	InvoiceCCEmail		string

	// Invoice amount in dollars
	TotalAmount		float64

	// Date the invoice is finalized
	InvoiceDate		time.Time

	// Purchase Order associated with the Invoice
	// FIXME: not found in output (Oct 2015)
	PoNumber		string

	// Usage details of a resource or collection of similar resources
	LineItems		[]struct{
		// Quantity of the line item
		Quantity	int

		// Text description of the line item
		Description	string

		// Unit cost of the line item
		UnitCost	float64

		// Total cost of the line item
		ItemTotal	float64

		// Location of the line item
		ServiceLocation	string

		// Individual line item description and cost.
		// For instance, may refer to the servers within a group.
		ItemDetails		[]struct {
			// Typically the name of the lowest level billed resource, such as a server name.
			Description	string

			// Cost of this resource.
			Cost		float64
		}
	}
}

// Get a list of invoicing data (estimates) for a given account alias for a given month.
// @year:           Year of usage, in YYYY format.
// @month:          Month of usage, in M{,M} format
// @date:           Date of the invoice data (only month and year are used)..
// @pricingAccount: Short code of the account that sends the invoice for the account alias.
func (c *Client) GetInvoiceData(year, month int, pricingAccount string) (res InvoiceData, err error) {
	path := fmt.Sprintf("/v2/invoice/%s/%4d/%d", c.AccountAlias, year, month)
	if pricingAccount != "" {
		path += "?pricingAccount=" + pricingAccount
	}
	err = c.getResponse("GET", path, nil, &res)
	return
}
