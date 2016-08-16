/*
 * Get the details of an individual group and any sub-groups and servers that it contains.
 */
package main

import (
	"github.com/olekukonko/tablewriter"
	"github.com/dustin/go-humanize"
	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
	"github.com/kr/pretty"
	"encoding/hex"
	"strings"
	"path"
	"flag"
	"fmt"
	"os"
)

func main() {
	var uuid string
	var simple   = flag.Bool("simple", false, "Use simple (debugging) output format")
	var location = flag.String("l", "", "Location to use if using a Group-Name instead of a UUID")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <Group Name or UUID>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	client, err := clcv2.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	if _, err := hex.DecodeString(flag.Arg(0)); err == nil {
		uuid = flag.Arg(0)
	} else if  *location == "" {
		exit.Errorf("Need a location argument (-l) if not using Group UUID (%s)", flag.Arg(0))
	} else {
		if grp, err := client.GetGroupByName(flag.Arg(0), *location); err != nil {
			exit.Errorf("failed to resolve group name %q: %s", flag.Arg(0), err)
		} else if grp == nil {
			exit.Errorf("No group named %q was found in %s", flag.Arg(0), *location)
		} else {
			uuid = grp.Id
		}
	}

	rootNode, err := client.GetGroup(uuid)
	if err != nil {
		exit.Fatalf("failed to query HW group %s: %s", flag.Arg(0), err)
	}

	if *simple {
		pretty.Println(rootNode)
	} else {
		fmt.Printf("Group %q in %s:\n", rootNode.Name, rootNode.LocationId)
		fmt.Printf("ID:            %s\n", rootNode.Id)
		fmt.Printf("Description:   %s\n", rootNode.Description)
		fmt.Printf("Type:          %s\n", rootNode.Type)
		fmt.Printf("Status:        %s\n", rootNode.Status)

		if len(rootNode.CustomFields) > 0 {
			fmt.Println("Custom fields:", rootNode.CustomFields)
		}

		// ChangeInfo
		createdStr := humanize.Time(rootNode.ChangeInfo.CreatedDate)
		/* The CreatedBy field can be an email address, or an API Key (hex string) */
		if _, err := hex.DecodeString(rootNode.ChangeInfo.CreatedBy); err == nil {
			createdStr += " via API Key"
		} else {
			createdStr += " by " + rootNode.ChangeInfo.CreatedBy
		}
		fmt.Printf("Created:       %s\n", createdStr)

		modifiedStr := humanize.Time(rootNode.ChangeInfo.ModifiedDate)
		/* The ModifiedBy field can be an email address, or an API Key (hex string) */
		if _, err := hex.DecodeString(rootNode.ChangeInfo.ModifiedBy); err == nil {
			modifiedStr += " via API Key"
		} else {
			modifiedStr += " by " + rootNode.ChangeInfo.ModifiedBy
		}
		fmt.Printf("Modified:      %s\n", modifiedStr)

		// Servers
		fmt.Printf("#Servers:      %d\n", rootNode.Serverscount)
		if rootNode.Serverscount > 0 {
			var servers []string

			if sl := clcv2.ExtractLinks(rootNode.Links, "server"); len(sl) > 0 {
				for _, s := range sl {
					servers = append(servers, s.Id)
				}
				fmt.Printf("Servers:       %s\n", strings.Join(servers, ", "))
			}
		}

		// Sub-groups
		if len(rootNode.Groups) > 0 {
			fmt.Printf("\nGroups of %s:\n", rootNode.Name)
			table := tablewriter.NewWriter(os.Stdout)
			table.SetAutoFormatHeaders(false)
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.SetAutoWrapText(true)

			table.SetHeader([]string{ "Name", "UUID", "Description", "#Servers", "Type"})
			for _, g := range rootNode.Groups {
				table.Append([]string{ g.Name, g.Id, g.Description, fmt.Sprint(g.Serverscount), g.Type })
			}
			table.Render()
		} else {
			fmt.Printf("Sub-groups:    none\n")
		}


	}
}
