/*
 * Lists all load balancers for a given data centre.
 */
package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
	"github.com/kr/pretty"
	"github.com/olekukonko/tablewriter"
)

func main() {
	var simple = flag.Bool("simple", false, "Use simple (debugging) output format")

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

	client, err := clcv2.NewClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	lbl, err := client.GetSharedLoadBalancers(flag.Arg(0))
	if err != nil {
		exit.Fatalf("Failed to list load balancers in %s: %s", flag.Arg(0), err)
	}

	if len(lbl) == 0 {
		fmt.Printf("No load balancers in %s.\n", flag.Arg(0))
	} else if *simple {
		pretty.Println(lbl)
	} else {
		fmt.Printf("Load balancers in %s:\n", flag.Arg(0))
		for _, lb := range lbl {
			fmt.Printf("Balancer:   %q (%s), ID %s\n", lb.Name, lb.Description, lb.ID)
			if lb.IpAddress != "" {
				fmt.Printf("IP Address: %s\n", lb.IpAddress)
			}
			if lb.Status != "" {
				fmt.Printf("Status:     %s\n", lb.Status)
			}
			fmt.Println("Pools:")

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
		}
	}
}
