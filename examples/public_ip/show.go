/*
 * Print details of a public IP address on a server.
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
)

func main() {
	var server string
	var publicIPs []string

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <Server-Name> [<Public-IP> ...]\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}
	server = flag.Arg(0)
	for i := 1; i < flag.NArg(); i++ {
		publicIPs = append(publicIPs, flag.Arg(i))
	}

	client, err := clcv2.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	if len(publicIPs) == 0 {
		srv, err := client.GetServer(server)
		if err != nil {
			exit.Fatalf("failed to query the public IPs of %s: %s", server, err)
		}

		for _, ip := range srv.Details.IpAddresses {
			if ip.IsPublic() {
				publicIPs = append(publicIPs, ip.Public)
			}
		}

		if len(publicIPs) == 0 {
			fmt.Printf("%s is not associated with any public IP address.\n", server)
			os.Exit(0)
		}
	}

	for _, ip := range publicIPs {
		p, err := client.GetPublicIPAddress(server, ip)
		if err != nil {
			exit.Fatalf("failed to query public IP address %s of %s: %s", ip, server, err)
		}

		fmt.Printf("%s:\n", server)
		fmt.Printf("Public IP:           %s\n", ip)
		fmt.Printf("Internal IP:         %s\n", p.InternalIPAddress)

		if len(p.Ports) > 0 {
			var ports []string

			for _, port := range p.Ports {
				ports = append(ports, port.String())
			}
			fmt.Printf("Port Openings:       %s\n", strings.Join(ports, ", "))
		}
		if len(p.SourceRestrictions) > 0 {
			var srcRes []string

			for _, src := range p.SourceRestrictions {
				srcRes = append(srcRes, src.Cidr)
			}

			fmt.Printf("Source Restrictions: %s\n", strings.Join(srcRes, ", "))
		}
	}
}
