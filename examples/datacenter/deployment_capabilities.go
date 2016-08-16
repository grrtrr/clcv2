/*
 * For a given (or default) datacenter, list its deployment capabilities.
 */
package main

import (
	"github.com/olekukonko/tablewriter"
	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
	"path"
	"flag"
	"fmt"
	"os"
)

func main() {
	var short = flag.Bool("s", false, "produce short output (names only)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  Location-Alias\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
		os.Exit(0)
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

	capa, err := client.GetDeploymentCapabilities(flag.Arg(0))
	if err != nil {
		exit.Fatalf("failed to query deployment capabilities of %s: %s", flag.Arg(0), err)
	}

	fmt.Printf("Datacenter %s:\n", flag.Arg(0))
	fmt.Printf( "Datacenter enabled:            %t\n", capa.DataCenterEnabled)
	fmt.Printf( "VM import enabled:             %t\n", capa.ImportVMEnabled)
	fmt.Println("Supports premium storage:     ", capa.SupportsPremiumStorage)
	fmt.Println("Supports shared load balancer:", capa.SupportsSharedLoadBalancer)
	fmt.Println("Supports bare metal servers:  ", capa.SupportsBareMetalServers)

	fmt.Println("\nDeployable Networks:")
	if len(capa.DeployableNetworks) == 0 {
		fmt.Println("\tNo deployable networks defined.")
	} else if *short {
		for _, net := range capa.DeployableNetworks {
			fmt.Printf("\t%s\n", net.Name)
		}
	} else {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoFormatHeaders(false)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetAutoWrapText(false)

		table.SetHeader([]string{ "Name", "Type", "Account", "Network ID" })
		for _, net := range capa.DeployableNetworks {
			table.Append([]string{ net.Name, net.Type, net.AccountID, net.NetworkId })
		}

		table.Render()
	}

	fmt.Println("\nAvailable Templates:")
	if *short {
		for _, tpl := range capa.Templates {
			fmt.Printf("\t%s\n", tpl.Name)
		}
	} else {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoFormatHeaders(false)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetAutoWrapText(false)

		/* Note: not displaying ReservedDrivePaths and DrivePathLength here, I don't understand their use. */
		/* Note: not listing Capabilities here, since the table gets too large for a single screen */
		table.SetHeader([]string{ "Name", "Description", "OS", "Storage" })

		for _, tpl := range capa.Templates {
			table.Append([]string{ tpl.Name, tpl.Description, tpl.OsType, fmt.Sprintf("%d GB", tpl.StorageSizeGB) })
		}
		table.Render()
	}

	fmt.Println("\nImportable OS Types:")
	if *short {
		for _, ios := range capa.ImportableOsTypes {
			fmt.Printf("\t%s\n", ios.Type)
		}
	} else {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoFormatHeaders(false)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetAutoWrapText(false)

		table.SetHeader([]string{ "Type", "Description", "Id", "Lab Product Code", "Premium Product Code" })

		for _, ios := range capa.ImportableOsTypes {
			table.Append([]string{ ios.Type, ios.Description, fmt.Sprintf("%d", ios.Id),
				ios.LabProductCode, ios.PremiumProductCode })
		}
		table.Render()
	}
}
