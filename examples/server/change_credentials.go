/*
 * Change the administrator/root password on an existing server given the current administrator/root password.
 */
package main

import (
	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
	"path"
	"flag"
	"fmt"
	"os"
)

func main() {
	var oldPasswd = flag.String("old", "", "The existing password (optional, used for authentication)")
	var newPasswd = flag.String("new", "", "The new password to apply")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <server-name>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() != 1 || *newPasswd == "" {
		flag.Usage()
		os.Exit(1)
	}

	client, err := clcv2.NewClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	if  *oldPasswd == "" {
		fmt.Printf("Looking up existing credentials of %s ...\n", flag.Arg(0))

		credentials, err := client.GetServerCredentials(flag.Arg(0))
		if err != nil {
			exit.Fatalf("Failed to obtain the credentials of server %q: %s", flag.Arg(0), err)
		}

		fmt.Printf("Existing %q password on %s: %q\n", credentials.Username,
			   flag.Arg(0), credentials.Password)
		*oldPasswd = credentials.Password
	}

	statusId, err := client.ServerChangePassword(flag.Arg(0), *oldPasswd, *newPasswd)
	if err != nil {
		exit.Fatalf("Failed to change the password on %q: %s", flag.Arg(0), err)
	}

	fmt.Printf("Status Id for changing the password on %s to %q: %s\n", flag.Arg(0), *newPasswd, statusId)
}
