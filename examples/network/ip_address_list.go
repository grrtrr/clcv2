/*
 * Lists IP addresses for a network in a given data center for a given account.
 */
package main

import (
	"github.com/olekukonko/tablewriter"
	"github.com/grrtrr/clcv2/utils"
	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
	"github.com/kr/pretty"
	"encoding/hex"
	"path"
	"flag"
	"net"
	"fmt"
	"os"
)

func main() {
	var netw string	/* The network ID, name or CIDR to query */
	var query    = flag.String("q", "free",   "Filter IP addresses; one of 'claimed', 'free', or 'all'")
	var location = flag.String("l", "",       "Data centre alias of the network")
	var simple   = flag.Bool("simple", false, "Use simple (debugging) output format")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <Network-ID|Network-Name|Network-CIDR>)>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	/* It seems that Location is always required, even if using the hex ID of the network. */
	if flag.NArg() != 1 || *location == "" {
		flag.Usage()
		os.Exit(1)
	} else if !inStringArray(*query, "claimed", "free", "all") {
		exit.Errorf("Invalid IP query %q. Try -h")
	}

	client, err := clcv2.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	/* Allow argument to be one of hex-ID, CIDR, or network name. */
	if _, err := hex.DecodeString(flag.Arg(0)); err == nil {
		netw = flag.Arg(0)
	} else if  *location == "" {
		exit.Errorf("Need a location argument (-l) if not using a network ID (%s)", flag.Arg(0))
	} else if _, _, err := net.ParseCIDR(flag.Arg(0)); err == nil {
		if network, err := client.GetNetworkIdByCIDR(flag.Arg(0), *location); err != nil {
			exit.Errorf("Failed to resolve network %s: %s", flag.Arg(0), err)
		} else if network == nil {
			exit.Errorf("No network of type %s was found in %s", flag.Arg(0), *location)
		} else {
			netw = network.Id
		}
	} else {
		if network, err := client.GetNetworkIdByName(flag.Arg(0), *location); err != nil {
			exit.Errorf("Failed to resolve network name %q: %s", flag.Arg(0), err)
		} else if network == nil {
			exit.Errorf("No network named %q was found in %s", flag.Arg(0), *location)
		} else {
			netw = network.Id
		}
	}

	// Note: re-using the 'Get Network' call here. There seems to be no dedicated 'Get IP Address List' call.
	details, err := client.GetNetworkDetails(*location, netw, *query)
	if err != nil {
		exit.Fatalf("Failed to query network details of %s: %s", netw, err)
	}

	if len(details.IpAddresses) == 0 {
		println("Empty result.")
	} else if *simple {
		pretty.Println(details)
	} else if *query == "free" {
		fmt.Printf("Free IP addresses on %s (%s):\n", details.Cidr, details.Name)
		for _, rng := range utils.CollapseIpRanges(clcv2.ExtractIPs(details.IpAddresses)) {
			fmt.Println(rng)
		}
	} else {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoFormatHeaders(false)
		table.SetAlignment(tablewriter.ALIGN_RIGHT)
		table.SetAutoWrapText(false)

		switch *query {
		case "claimed":
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.SetHeader([]string{ "Address", "Server" })
			for _, i := range details.IpAddresses {
				table.Append([]string{ i.Address, i.Server })
			}
		case "all":
			table.SetHeader([]string{ "Address", "Claimed", "Server", "Type" })
			for _, i := range details.IpAddresses {
				table.Append([]string{ i.Address, fmt.Sprint(i.Claimed), i.Server, i.Type })
			}
		}
		table.Render()
	}
}

/* go replacement for python 'x in list' */
func inStringArray(s string, list ...string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
