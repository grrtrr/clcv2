/*
 * Create a new intra-datacenter firewall policy.
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
	var src, dst clcv2.CIDRs
	var ports clcv2.PortSpecs

	var acct = flag.String("da", "", "Destination account (defaults to source account)")
	flag.Var(&src, "src", "Source network(s) in CIDR notation (option may be repeated)")
	flag.Var(&dst, "dst", "Destination network(s) in CIDR notation (option may be repeated)")
	flag.Var(&ports, "p", "Port spec(s), number(s) or service name(s) (option may be repeated)\n"+
		"        - ping:      use ping or icmp\n"+
		"        - full spec: tcp/20081-20083, udp/554, udp/6080-7000, ...\n"+
		"        - tcp names: rdp, http, https, http-alt, ssh, ftp, ftps, ...\n"+
		"        - tcp ports: 22, 443, 80, 20081-20083, ...\n"+
		"        - DEFAULTS:  ping, ssh, http")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options] -src <SrcCIDR> -dst <DstCIDR>  <Location>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() != 1 || len(src) == 0 || len(dst) == 0 {
		flag.Usage()
		os.Exit(0)
	}

	client, err := clcv2.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	if *acct == "" {
		*acct = client.AccountAlias
	}

	req := clcv2.IntraDataCenterFirewallPolicyReq{
		SourceCIDR:  src,
		DestCIDR:    dst,
		DestAccount: *acct,
		Ports:       clcv2.PortSpecString(ports),
	}

	id, err := client.CreateIntraDataCenterFirewallPolicy(flag.Arg(0), &req)
	if err != nil {
		exit.Fatalf("failed to create intra-datacenter firewall policy in %s: %s", flag.Arg(0), err)
	}
	fmt.Printf("Successfully created intra-datacenter firewall policy %s in %s.\n", id, strings.ToUpper(flag.Arg(0)))
}
