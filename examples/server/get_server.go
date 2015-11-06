/*
 * List the details of one server
 */
package main

import (
	"github.com/olekukonko/tablewriter"
	"github.com/dustin/go-humanize"
	"github.com/grrtrr/clcv2/utils"
	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
	"encoding/hex"
	"strings"
	"path"
	"flag"
	"log"
	"fmt"
	"os"
)

func main() {
	var simple = flag.Bool("simple", false, "Use simple (debugging) output format")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <Server Name>\n", path.Base(os.Args[0]))
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

	server, err := client.GetServer(flag.Arg(0))
	if err != nil {
		exit.Fatalf("Failed to list details of server %q: %s", flag.Arg(0), err)
	}

	if *simple {
		utils.PrintStruct(server.Details)
	} else {
		grp, err := client.GetGroup(server.GroupId)
		if err != nil {
			exit.Fatalf("Failed to resolve group UUID: %s", err)
		}

		/* First public, then private */
		IPs := []string{}
		for _, ip := range server.Details.IpAddresses {
			if ip.Public != "" {
				IPs = append(IPs, ip.Public)
			}
		}
		for _, ip := range server.Details.IpAddresses {
			if ip.Internal != "" {
				IPs = append(IPs, ip.Internal)
			}

		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoFormatHeaders(false)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetAutoWrapText(true)

		table.SetHeader([]string{
			"Name", "Group", "Description", "OS",
			"CPU", "Mem/GB",
			"IP", "Power", "Last Change",
		})

		modifiedStr := humanize.Time(server.ChangeInfo.ModifiedDate)
		/* The ModifiedBy field can be an email address, or an API Key (hex string) */
		if _, err := hex.DecodeString(server.ChangeInfo.ModifiedBy); err == nil {
			modifiedStr += " via API Key"
		} else if len(server.ChangeInfo.ModifiedBy) > 6 {
			modifiedStr += " by " + server.ChangeInfo.ModifiedBy[:6]
		} else {
			modifiedStr += " by " + server.ChangeInfo.ModifiedBy
		}

		table.Append([]string{
			server.Name, grp.Name, server.Description, server.OsType,
			fmt.Sprint(server.Details.Cpu),	fmt.Sprintf("%d", server.Details.MemoryMb/1024),
			strings.Join(IPs, " "),
			server.Details.PowerState, modifiedStr,
		})
		table.Render()


		// Disks
		fmt.Printf("Disks of %s (total storage: %d GB)\n", server.Name, server.Details.StorageGb)
		table = tablewriter.NewWriter(os.Stdout)
		table.SetAutoFormatHeaders(false)
		table.SetAlignment(tablewriter.ALIGN_RIGHT)
		table.SetAutoWrapText(true)

		table.SetHeader([]string{ "Disk ID", "Disk Size/GB", "Paths" })
		for _, d := range server.Details.Disks {
			table.Append([]string{ d.Id, fmt.Sprint(d.SizeGb), strings.Join(d.PartitionPaths, ", ")})
		}
		table.Render()

		// Partitions
		fmt.Printf("Disks of %s:\n", server.Name)
		table = tablewriter.NewWriter(os.Stdout)
		table.SetAutoFormatHeaders(false)
		table.SetAlignment(tablewriter.ALIGN_RIGHT)
		table.SetAutoWrapText(true)

		table.SetHeader([]string{ "Partition Path", "Partition Size/GB" })
		for _, p := range server.Details.Partitions {
			table.Append([]string{ p.Path, fmt.Sprintf("%.1f", p.SizeGb) })
		}
		table.Render()
	}
}
