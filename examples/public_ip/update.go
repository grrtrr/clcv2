/*
 * (Re)configure firewall/port settings of a given public IP Address.
 */
package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
)

func main() {
	var srcRes clcv2.SrcRestrictions
	var portSp clcv2.PortSpecs
	flag.Var(&srcRes, "src", "Restrict source traffic to given CIDR range(s)")
	flag.Var(&portSp, "p", "Port spec(s), number(s) or service name(s)\n"+
		"        - ping:      use ping or icmp\n"+
		"        - full spec: tcp/20081-20083, udp/554, udp/6080-7000, ...\n"+
		"        - tcp names: rdp, http, https, http-alt, ssh, ftp, ftps, ...\n"+
		"        - tcp ports: 22, 443, 80, 20081-20083, ...")
	var keep = flag.Bool("k", true, "Keep the existing settings")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <server-name> <public-IP>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() != 2 || len(portSp) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	client, err := clcv2.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	if *keep {
		fmt.Printf("Looking up existing configuration of %s on %s ...\n", flag.Arg(1), flag.Arg(0))
		old, err := client.GetPublicIPAddress(flag.Arg(0), flag.Arg(1))
		if err != nil {
			exit.Fatalf("failed to get existing configuration for %s: %s", flag.Arg(1), err)
		}
		log.Printf("existing settings: %v", old.Ports)
		portSp = append(portSp, old.Ports...)
	}

	req := clcv2.PublicIPAddress{Ports: portSp, SourceRestrictions: srcRes}
	reqId, err := client.UpdatePublicIPAddress(flag.Arg(0), flag.Arg(1), &req)
	if err != nil {
		exit.Fatalf("failed to update public IP address %s on %q: %s", flag.Arg(1), flag.Arg(0), err)
	}

	client.PollStatus(reqId, 1*time.Second)
}
