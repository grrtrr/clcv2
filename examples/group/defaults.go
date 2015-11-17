/*
 * Set the defaults for a Hardware Group.
 */
package main

import (
	"github.com/olekukonko/tablewriter"
	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
	"encoding/hex"
	"strings"
	"path"
	"flag"
	"fmt"
	"os"
)

func main() {
	var group string	/* UUID of the group to set defaults of */
	var location = flag.String("l",   "",  "Location to use if using a Group-Name instead of a UUID")
	var template = flag.String("t",   "",          "Name of the template to use as the source")
	var numCpu   = flag.Int("cpu",     1,          "Number of Cpus to use (1-16)")
	var memGB    = flag.Int("mem",     4,          "Amount of memory in GB (1-128)")
	var net      = flag.String("net",  "",         "Name of the Network to use")
	var primDNS  = flag.String("dns1", "8.8.8.8",  "Primary DNS to use")
	var secDNS   = flag.String("dns2", "8.8.4.4",  "Secondary DNS to use")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <Group Name or UUID>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(0)
	}

	client, err := clcv2.NewClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	if _, err := hex.DecodeString(flag.Arg(0)); err == nil {
		group = flag.Arg(0)
	} else if  *location == "" {
		exit.Errorf("Need a location argument (-l) if not using Group UUID (%s)", flag.Arg(0))
	} else {
		fmt.Printf("Resolving group id of %q ...\n", flag.Arg(0))
		if grp, err := client.GetGroupByName(flag.Arg(0), *location); err != nil {
			exit.Errorf("Failed to resolve group name %q: %s", flag.Arg(0), err)
		} else if grp == nil {
			exit.Errorf("No group named %q was found on %s", flag.Arg(0), *location)
		} else {
			group = grp.Id
		}
	}
	if *net != "" {
		if _, err := hex.DecodeString(*net); err == nil {
			/* already looks like a HEX ID */
		} else if  *location == "" {
			exit.Errorf("Need a location argument (-l) if not using a network ID (%s)", *net)
		} else {
			fmt.Printf("Resolving network id of %q ...\n", *net)

			if netw, err := client.GetNetworkIdByName(*net, *location); err != nil {
				exit.Errorf("Failed to resolve network name %q: %s", *net, err)
			} else if netw == nil {
				exit.Errorf("No network named %q was found on %s", *net, *location)
			} else {
				*net = netw.Id
			}
		}
	}

	req := clcv2.GroupDefaults{
		Cpu:          *numCpu,
		MemoryGB:     *memGB,
		NetworkId:    *net,
		PrimaryDns:   *primDNS,
		SecondaryDns: *secDNS,
		TemplateName: *template,
	}

	settings, err := client.SetGroupDefaults(group, &req)
	if err != nil {
		exit.Fatalf("Failed to set group defaults of %s: %s", flag.Arg(0), err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetAlignment(tablewriter.ALIGN_RIGHT)
	table.SetAutoWrapText(false)

	table.SetHeader([]string{ "Default", "Value", "Inherited" })
	for k, v := range settings {
		table.Append([]string{ strings.Title(k), fmt.Sprint(v.Value), fmt.Sprint(v.Inherited) })
	}
	table.Render()
}
