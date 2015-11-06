/*
 * Retrieve the administrator/root password on an existing server.
 */
package main

import (
	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
	"path"
	"flag"
	"log"
	"fmt"
	"os"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <server-name>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	client, err := clcv2.NewClient(nil, log.New(os.Stdout, "", log.LstdFlags | log.Ltime))
	if err != nil {
		exit.Fatal(err.Error())
	}

	credentials, err := client.GetServerCredentials(flag.Arg(0))
	if err != nil {
		exit.Fatalf("Failed to obtain the credentials of server %q: %s", flag.Arg(0), err)
	}

	fmt.Printf("Credentials for %s:\n", flag.Arg(0))
	fmt.Printf("User:     %s\n", credentials.Username)
	fmt.Printf("Password: %s\n", credentials.Password)
}
