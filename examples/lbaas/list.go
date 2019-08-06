/*
 * Lists all LBaaS instances of a given data centre.
 */
package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/grrtrr/clcv2/clcv2cli"
	"github.com/grrtrr/exit"
)

func main() {
	var location string

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <Location>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()

	/* The Location argument is always required */
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}
	location = strings.ToUpper(flag.Arg(0))

	client, err := clcv2cli.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	lbl, err := client.GetLbInstances(location)
	if err != nil {
		exit.Fatalf("failed to list load balancers in %s: %s", location, err)
	}

	if len(lbl) == 0 {
		fmt.Printf("No load balancers in %s.\n", location)
	} else {
		fmt.Printf("Load balancers in %s:\n", location)
		for _, lb := range lbl {
			fmt.Printf("Balancer:   %s (%q, %q)\n", lb.ID, lb.Name, lb.Description)
			fmt.Printf("Owner:      %s @ %s\n", lb.Owner, strings.ToUpper(lb.DataCenter))
			fmt.Printf("IP Address: %s\n", lb.PublicIP)
			fmt.Printf("Created:    %s\n", lb.Created.Time().Format(`Mon 2 Jan 2006, 15:04:05 MST`))
			if lb.Deleted != nil {
				fmt.Printf("Deleted:    %s\n", lb.Deleted.Time().Format(`Mon 2 Jan 2006, 15:04:05 MST`))
			}
			if lb.Status != "" {
				fmt.Printf("Status:     %s\n", lb.Status)
			}

			fmt.Println("Pools:")
			/*
				for _, p := range lb.Pools {
					var nodes []string

					table := tablewriter.NewWriter(os.Stdout)
					table.SetAutoFormatHeaders(false)
					table.SetAlignment(tablewriter.ALIGN_LEFT)
					table.SetAutoWrapText(true)

					table.Append([]string{"Pool ID", p.ID})
					table.Append([]string{"Port", fmt.Sprint(p.Port)})
					table.Append([]string{"Method", p.Method})
					table.Append([]string{"Persistence", p.Persistence})

					for _, n := range p.Nodes {
						var s = "???"
						switch n.Status {
						case "disabled":
							s = "-"
						case "enabled":
							s = ""
						case "deleted":
							s = "~"
						}
						s += fmt.Sprintf("%s:%d", n.IPAddress, n.PrivatePort)
						nodes = append(nodes, s)
					}
					table.Append([]string{"Nodes", strings.Join(nodes, " ")})
					table.Render()
				}
			*/

		}
	}
}
