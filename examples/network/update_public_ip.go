/*
 * Configure firewall settings of a public IP Address.
 */
package main

import (
	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
	"path"
	"flag"
	"fmt"
	"os"
)

func main() {
	var http     = flag.Bool("http",    false, "Allow HTTP requests (port 80) on the new IP")
	var http8080 = flag.Bool("httpAlt", false, "Allow HTTP requests (port 8080) on the new IP")
	var https    = flag.Bool("https",   false, "Allow HTTPS requests (port 443) on the new IP")
	var ftp      = flag.Bool("ftp",     false, "Allow FTP requests (port 21) on the new IP")
	var ftps     = flag.Bool("ftps",    false, "Allow FTPS requests (port 990) on the new IP")
	var ssh      = flag.Bool("ssh",     true,  "Allow SSH requests (port 22) on the new IP")
	var sftp     = flag.Bool("sftp",    true,  "Allow SFTP requests (port 22) on the new IP")
	var rdp      = flag.Bool("rdp",     false, "Allow RDP requests (port 3389) on the new IP")
	var src      = flag.String("src",   "",    "Restrict source traffic to this CIDR range")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  <server-name> <public-IP>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(1)
	}

	client, err := clcv2.NewClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	req := clcv2.PublicIPAddress{}
	if *http {
		req.Ports = append(req.Ports, clcv2.PublicPort{ "tcp", 80, 0 })
	}
	if *http8080 {
		req.Ports = append(req.Ports, clcv2.PublicPort{ "tcp", 8080, 0 })
	}
	if *https {
		req.Ports = append(req.Ports, clcv2.PublicPort{ "tcp", 443, 0 })
	}
	if *ftp {
		req.Ports = append(req.Ports, clcv2.PublicPort{ "tcp", 21, 0 })
	}
	if *ftps {
		req.Ports = append(req.Ports, clcv2.PublicPort{ "tcp", 990, 0 })
	}
	if *ssh || *sftp {
		req.Ports = append(req.Ports, clcv2.PublicPort{ "tcp", 22, 0 })
	}
	if *rdp {
		req.Ports = append(req.Ports, clcv2.PublicPort{ "tcp", 3389, 0 })
	}

	if *src != "" {
		req.SourceRestrictions = append(req.SourceRestrictions, clcv2.SourceCIDR{*src})
	}

	reqId, err := client.UpdatePublicIPAddress(flag.Arg(0), flag.Arg(1), &req)
	if err != nil {
		exit.Fatalf("Failed to update public IP address on %q: %s", flag.Arg(0), err)
	}

	fmt.Println("Request ID for adding public IP:", reqId)
}
