/*
 * Add a public IP address to a server
 */
package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/clcv2/clcv2cli"
	"github.com/grrtrr/exit"
)

func main() {
	var srcRes clcv2.SrcRestrictions
	var portSp clcv2.PortSpecs

	flag.Var(&srcRes, "src", "Restrict source traffic to given CIDR range(s)")
	flag.Var(&portSp, "p", "Port spec(s), number(s) or service name(s) (option can be repeated)\n"+
		"        - ping:      use ping or icmp\n"+
		"        - full spec: tcp/20081-20083, udp/554, udp/6080-7000, ...\n"+
		"        - tcp names: rdp, http, https, http-alt, ssh, ftp, ftps, ...\n"+
		"        - tcp ports: 22, 443, 80, 20081-20083, ...\n"+
		"        - DEFAULTS:  ping, ssh, http")
	var ipAddr = flag.String("i", "", "Use this existing internal IP on the server")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <server-name>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	} else if len(portSp) == 0 { /* default port settings */
		portSp.Set("ping")
		portSp.Set("ssh")
		portSp.Set("http")
	}

	client, err := clcv2cli.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	req := clcv2.PublicIPAddress{
		InternalIPAddress:  *ipAddr,
		Ports:              portSp,
		SourceRestrictions: srcRes,
	}

	reqId, err := client.AddPublicIPAddress(flag.Arg(0), &req)
	if err != nil {
		exit.Fatalf("failed to add a public IP address to %q: %s", flag.Arg(0), err)
	}

	client.PollStatus(reqId, 1*time.Second)
}
