/*
 * Get a list of invoicing data (estimates) for a given account alias for a given month.
 */
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/grrtrr/clcv2/clcv2cli"
	"github.com/grrtrr/exit"
	"github.com/olekukonko/tablewriter"
)

func main() {
	var now = time.Now()
	var pricingAcct = flag.String("pricing", "", "Pricing account (to use instead of default account alias)")
	var invoiceYear = flag.Int("y", now.Year(), "Year of the invoice date")
	var invoiceMonth = flag.Int("m", int(now.Month())-1, "Month of the invoice date")
	var itemDetails = flag.Bool("details", false, "Print individual line item details also")

	flag.Parse()

	client, err := clcv2cli.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	id, err := client.GetInvoiceData(*invoiceYear, *invoiceMonth, *pricingAcct)
	if err != nil {
		exit.Fatalf("failed to obtain invoice data: %s", err)
	}

	fmt.Printf("Details of invoice %q for %s (%s):\n", id.Id, id.PricingAccountAlias, id.CompanyName)
	fmt.Printf("Address:               %s, %s, %s, %s\n", id.Address1, id.City, id.StateProvince, id.PostalCode)
	fmt.Printf("Parent account alias:  %s\n", id.ParentAccountAlias)
	fmt.Printf("Billing contact email: %s\n", id.BillingContactEmail)
	fmt.Printf("Invoice date:          %s\n", id.InvoiceDate.Format("January 2006"))
	fmt.Printf("Terms:                 %s\n", id.Terms)
	fmt.Printf("Total amount:          $%.2f\n", id.TotalAmount)

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetAutoWrapText(false)
	table.SetAlignment(tablewriter.ALIGN_RIGHT)

	table.SetHeader([]string{
		"Description", "Location", "Quantity", "Unit Cost", "Total",
	})

	for _, li := range id.LineItems {
		table.Append([]string{
			li.Description, li.ServiceLocation, fmt.Sprint(li.Quantity),
			fmt.Sprintf("%.2f", li.UnitCost), fmt.Sprintf("%.2f", li.ItemTotal),
		})
	}
	table.Render()

	if *itemDetails {
		fmt.Println("\n\nIndividual details:")
		for _, li := range id.LineItems {
			if len(li.ItemDetails) != 0 {
				fmt.Printf("%s ($%.2f total):\n", li.Description, li.ItemTotal)
				for _, det := range li.ItemDetails {
					fmt.Printf("\t%-20.20s $%.2f\n", det.Description, det.Cost)
				}
			}
		}
	}
}
