/*
 * List the AccountCustomFields associated with an account.
 * This will use the default Account Alias if -a <acctAlias> is not set.
 */
package main

import (
	"flag"

	"github.com/grrtrr/clcv2/clcv2cli"
	"github.com/grrtrr/clcv2/utils"
	"github.com/grrtrr/exit"
)

func main() {
	flag.Parse()

	client, err := clcv2cli.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	customFields, err := client.GetCustomFields()
	if err != nil {
		exit.Fatalf("failed to obtain Custom Fields: %s", err)
	}

	if len(customFields) == 0 {
		println("Empty result.")
	} else {
		for _, cf := range customFields {
			utils.PrintStruct(cf)
		}
	}
}
