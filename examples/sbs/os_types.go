/*
 * Print the list of Operating System Types supported by the Simple Backup Service.
 */
package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
)

func main() {
	flag.Parse()

	client, err := clcv2.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	osTypes, err := client.SBSgetOsTypes()
	if err != nil {
		exit.Fatalf("failed to retrieve the list of OS Types supported by SBS: %s", err)
	}
	fmt.Printf("Supported SBS OS Types: %s\n", strings.Join(osTypes, " and "))
}
