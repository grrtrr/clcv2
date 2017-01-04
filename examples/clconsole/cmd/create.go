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

var createFlags struct {
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
}

func init() {
	Create.Flags().StringVarP(&createFlags.hwGroup, "group", "g", "", "UUID or name (if unique) of the HW group to add this server to")
	Create.Flags().StringVar(&createFlags.srcServer, "src", "", "The name of a source-server, or a template, to create from")
	Create.Flags().StringVar(&createFlags.srcPass, "createFlags.srcPass", "", "When cloning from a source-server, use this createFlags.password")
	Create.Flags().StringVarP(&createFlags.seed, "createFlags.seed", "s", "AUTO", "The createFlags.seed for the server name")
	Create.Flags().StringVar(&createFlags.desc, "createFlags.desc", "", "Textual description of the server")

	Create.Flags().StringVar(&createFlags.net, "createFlags.net", "", "ID or name of the Network to use")
	Create.Flags().StringVar(&createFlags.primDNS, "dns1", "8.8.8.8", "Primary DNS to use")
	Create.Flags().StringVar(&createFlags.secDNS, "dns2", "8.8.4.4", "Secondary DNS to use")

	Create.Flags().StringVar(&createFlags.password, "pass", "", "Desired createFlags.password. Leave blank to auto-generate")
	Create.Flags().StringVar(&createFlags.serverType, "type", "standard", "The type of server to create (standard, hyperscale, or bareMetal)")

	Create.Flags().Uint8Var(&createFlags.numCpu, "cpu", 1, "Number of Cpus to use")
	Create.Flags().Uint32Var(&createFlags.memGB, "memory", 4, "Amount of memory in GB")
	Create.Flags().Uint32Var(&createFlags.extraDrv, "drive", 0, "Extra storage (in GB) to add to server as a raw disk")

	Create.Flags().DurationVar(&createFlags.ttl, "createFlags.ttl", 0, "Time span (counting from time of creation) until server gets deleted")

	Root.AddCommand(Create)
}

var Create = &cobra.Command{
	Use:   "create",
	Short: "Create a new server from a template",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if createFlags.hwGroup == "" {
			return errors.Errorf("Need a hardware group to create the server in (--group)")
		} else if createFlags.srcServer == "" {
			return errors.Errorf("Need a source server or template ID (--src)")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// createFlags.hwGroup may be hex uuid or group name
		if _, err := hex.DecodeString(createFlags.hwGroup); err != nil {
			log.Printf("Resolving ID of Hardware Group %q ...", createFlags.hwGroup)

			if group, err := client.GetGroupByName(createFlags.hwGroup, location); err != nil {
				log.Fatalf("failed to resolve group name %q: %s", createFlags.hwGroup, err)
			} else if group == nil {
				log.Fatalf("no group named %q was found in %s", createFlags.hwGroup, location)
			} else {
				createFlags.hwGroup = group.Id
			}
		}

		// createFlags.net is supposed to be a (hex) ID, but allow network names, too
		if createFlags.net != "" {
			if _, err := hex.DecodeString(createFlags.net); err == nil {
				/* already looks like a HEX ID */
			} else if location == "" {
				log.Fatalf("Need a location argument (-l) if not using a network ID (%s)", createFlags.net)
			} else {
				log.Printf("resolving network id of %q ...", createFlags.net)

				if netw, err := client.GetNetworkIdByName(createFlags.net, location); err != nil {
					log.Fatalf("failed to resolve network name %q: %s", createFlags.net, err)
				} else if netw == nil {
					log.Fatalf("No network named %q was found in %s", createFlags.net, location)
				} else {
					createFlags.net = netw.Id
				}
			}
		}

		req := clcv2.CreateServerReq{
			// Name of the server to create. Alphanumeric characters and dashes only.
			Name: createFlags.seed,

			// User-defined description of this server
			Description: createFlags.desc,

			// ID of the parent HW group.
			GroupId: createFlags.hwGroup,

			// ID of the server to use a source. May be the ID of a createFlags.srcServer, or when cloning, an existing server ID.
			SourceServerId: createFlags.srcServer,

			// The primary DNS to set on the server
			PrimaryDns: createFlags.primDNS,

			// The secondary DNS to set on the server
			SecondaryDns: createFlags.secDNS,

			// ID of the network to which to deploy the server.
			NetworkId: createFlags.net,

			// Password of administrator or root user on server.
			Password: createFlags.password,

			// Password of the source server, used only when creating a clone from an existing server.
			SourceServerPassword: createFlags.srcPass,

			// Number of processors to configure the server with (1-16)
			Cpu: int(createFlags.numCpu),

			// Number of GB of memory to configure the server with (1-128)
			MemoryGB: int(createFlags.memGB),

			// Whether to create a 'standard', 'hyperscale', or 'bareMetal' server
			Type: createFlags.serverType,

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

		if createFlags.extraDrv != 0 {
			req.AdditionalDisks = append(req.AdditionalDisks,
				clcv2.ServerAdditionalDisk{SizeGB: createFlags.extraDrv, Type: "raw"})
		}

		// Date/time that the server should be deleted.
		if createFlags.ttl != 0 {
			req.Ttl = new(time.Time)
			*req.Ttl = time.Now().Add(createFlags.ttl)
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
