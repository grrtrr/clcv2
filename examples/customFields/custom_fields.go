/*
 * List the AccountCustomFields associated with an account.
 */
package main

import (
	"github.com/grrtrr/clcv2/utils"
	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
	"flag"
	"log"
	"os"
)

func main() {
	var acctAlias = flag.String("a", "", "Account alias to use")

	flag.Parse()

	client, err := clcv2.NewClient(nil, log.New(os.Stdout, "", log.LstdFlags | log.Ltime))
	if err != nil {
		exit.Fatal(err.Error())
	}

	customFields, err := client.GetCustomFields(*acctAlias)
	if err != nil {
		exit.Fatalf("Failed to obtain Custom Fields: %s", err)
	}

	if len(customFields) == 0 {
		println("Empty result.")
	} else {
		for _, cf := range customFields {
			utils.PrintStruct(cf)
		}
	}
}
