/*
 * Create a new server. Does not support creation of bare-metal servers (yet).
 */
package cmd

import (
	"encoding/hex"
	"log"
	"time"

	"github.com/grrtrr/clcv2"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// Flags
var (
	hwGroup    string        // hardware group to create server in
	srcServer  string        // source server/template to create from
	srcPass    string        // when using a source-server, use this password
	seed       string        // 4-6 character name seed for the server name
	desc       string        // description of the server
	primDNS    string        // primary DNS
	secDNS     string        // secondary DNS
	net        string        // ID or name of the network to use
	password   string        // desired password to use
	serverType string        // server type: standard, hyperscale, or bareMetal
	numCpu     uint8         // number of CPU cores to use
	memGB      uint32        // amount of memory in GB
	extraDrv   uint32        // extra amount of storage in GB
	ttl        time.Duration // time span (counting from time of creation) until server gets deleted
)

func init() {
	Create.Flags().StringVarP(&hwGroup, "group", "g", "", "UUID or name (if unique) of the HW group to add this server to")
	Create.Flags().StringVar(&srcServer, "src", "", "The name of a source-server, or a template, to create from")
	Create.Flags().StringVar(&srcPass, "srcPass", "", "When cloning from a source-server, use this password")
	Create.Flags().StringVarP(&seed, "seed", "s", "AUTO", "The seed for the server name")
	Create.Flags().StringVar(&desc, "desc", "", "Textual description of the server")

	Create.Flags().StringVar(&net, "net", "", "ID or name of the Network to use")
	Create.Flags().StringVar(&primDNS, "dns1", "8.8.8.8", "Primary DNS to use")
	Create.Flags().StringVar(&secDNS, "dns2", "8.8.4.4", "Secondary DNS to use")

	Create.Flags().StringVar(&password, "pass", "", "Desired password. Leave blank to auto-generate")
	Create.Flags().StringVar(&serverType, "type", "standard", "The type of server to create (standard, hyperscale, or bareMetal)")

	Create.Flags().Uint8Var(&numCpu, "cpu", 1, "Number of Cpus to use")
	Create.Flags().Uint32Var(&memGB, "memory", 4, "Amount of memory in GB")
	Create.Flags().Uint32Var(&extraDrv, "drive", 0, "Extra storage (in GB) to add to server as a raw disk")

	Create.Flags().DurationVar(&ttl, "ttl", 0, "Time span (counting from time of creation) until server gets deleted")

	Root.AddCommand(Create)
}

var Create = &cobra.Command{
	Use:   "create",
	Short: "Create a new server from a template",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if hwGroup == "" {
			return errors.Errorf("Need a hardware group to create the server in (--group)")
		} else if srcServer == "" {
			return errors.Errorf("Need a source server or template ID (--src)")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// hwGroup may be hex uuid or group name
		if _, err := hex.DecodeString(hwGroup); err != nil {
			log.Printf("Resolving ID of Hardware Group %q ...", hwGroup)

			if group, err := client.GetGroupByName(hwGroup, location); err != nil {
				log.Fatalf("failed to resolve group name %q: %s", hwGroup, err)
			} else if group == nil {
				log.Fatalf("no group named %q was found in %s", hwGroup, location)
			} else {
				hwGroup = group.Id
			}
		}

		// net is supposed to be a (hex) ID, but allow network names, too
		if net != "" {
			if _, err := hex.DecodeString(net); err == nil {
				/* already looks like a HEX ID */
			} else if location == "" {
				log.Fatalf("Need a location argument (-l) if not using a network ID (%s)", net)
			} else {
				log.Printf("resolving network id of %q ...", net)

				if netw, err := client.GetNetworkIdByName(net, location); err != nil {
					log.Fatalf("failed to resolve network name %q: %s", net, err)
				} else if netw == nil {
					log.Fatalf("No network named %q was found in %s", net, location)
				} else {
					net = netw.Id
				}
			}
		}

		req := clcv2.CreateServerReq{
			// Name of the server to create. Alphanumeric characters and dashes only.
			Name: seed,

			// User-defined description of this server
			Description: desc,

			// ID of the parent HW group.
			GroupId: hwGroup,

			// ID of the server to use a source. May be the ID of a srcServer, or when cloning, an existing server ID.
			SourceServerId: srcServer,

			// The primary DNS to set on the server
			PrimaryDns: primDNS,

			// The secondary DNS to set on the server
			SecondaryDns: secDNS,

			// ID of the network to which to deploy the server.
			NetworkId: net,

			// Password of administrator or root user on server.
			Password: password,

			// Password of the source server, used only when creating a clone from an existing server.
			SourceServerPassword: srcPass,

			// Number of processors to configure the server with (1-16)
			Cpu: int(numCpu),

			// Number of GB of memory to configure the server with (1-128)
			MemoryGB: int(memGB),

			// Whether to create a 'standard', 'hyperscale', or 'bareMetal' server
			Type: serverType,

			// FIXME: the following are not populated in this request:
			// - IpAddress
			// - IsManagedOs
			// - IsManagedBackup
			// - AntiAffinityPolicyId
			// - CpuAutoscalePolicyId
			// - CustomFields
			// - Packages
			//
			// The following items relevant specific to bare-metal servers are also ignored:
			// - ConfigurationId
			// - OsType
		}

		if extraDrv != 0 {
			req.AdditionalDisks = append(req.AdditionalDisks,
				clcv2.ServerAdditionalDisk{SizeGB: extraDrv, Type: "raw"})
		}

		// Date/time that the server should be deleted.
		if ttl != 0 {
			req.Ttl = new(time.Time)
			*req.Ttl = time.Now().Add(ttl)
		}

		// The CreateServer request resolves the server name at the end.
		// This second call can fail at the remote end; it does not mean that
		// the server has not been created yet.
		url, reqID, err := client.CreateServer(&req)
		if err != nil {
			log.Fatalf("failed to create server: %s", err)
		}
		log.Printf("Status Id: %s", reqID)
		client.PollStatus(reqID, 5*time.Second)

		// Print details after job completes
		server, err := client.GetServerByURI(url)
		if err != nil {
			log.Fatalf("failed to query server details at %s: %s", url, err)
		}
		showServer(client, server)
	},
}
