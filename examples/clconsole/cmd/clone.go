/*
 * Simplified interface for cloning an existing server.
 */
package cmd

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// cloneFlags contains flags similar to createFlags, which is why both are separate structs
var cloneFlags struct {
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
	Clone.Flags().StringVar(&cloneFlags.net, "net", "", "ID or name of the Network to use (if different from source)")
	Clone.Flags().StringVar(&cloneFlags.primDNS, "dns1", "8.8.8.8", "Primary DNS to use")
	Clone.Flags().StringVar(&cloneFlags.secDNS, "dns2", "8.8.4.4", "Secondary DNS to use")
	Clone.Flags().StringVarP(&cloneFlags.seed, "seed", "s", "", "The 4-6 character seed for the name of the cloned server")
	Clone.Flags().Uint8Var(&cloneFlags.numCpu, "cpu", 0, "Number of CPU cores to use (if different from source VM)")
	Clone.Flags().Uint32Var(&cloneFlags.memGB, "memory", 0, "Amount of memory in GB (if different from source VM")
	Clone.Flags().StringVar(&cloneFlags.desc, "desc", "", "Description of the cloned server")
	Clone.Flags().Uint32Var(&cloneFlags.extraDrv, "drive", 0, "Extra storage (in GB) to add to server as a raw disk")
	Clone.Flags().DurationVar(&cloneFlags.ttl, "ttl", 0, "Time span (counting from time of creation) until server gets deleted")

	Root.AddCommand(Clone)
}

var Clone = &cobra.Command{
	Use:     "clone  <srcName>  <destination>",
	Short:   "Clone existing server",
	Long:    "Clone source server @srcName and put it into the @destination folder",
	PreRunE: checkArgs(2, "Need a source server and a destination folder"),
	Run: func(cmd *cobra.Command, args []string) {
		var (
			source      = args[0]                         // source server
			dest        = strings.TrimRight(args[1], "/") // destination folder
			wasStopped  bool
			maxAttempts = 1
			url, reqID  string
		)

		// First get the details of the source server
		log.Printf("Obtaining details of source server %s ...", source)
		src, err := client.GetServer(source)
		if err != nil {
			exit.Fatalf("failed to list details of source server %q: %s", source, err)
		}

		if wasStopped = src.Details.PowerState == "stopped"; wasStopped {
			// The source server must be powered on
			log.Printf("%s is powered-off - powering on ...", src.Name)
			reqID, err := client.PowerOnServer(src.Name)
			if err != nil {
				exit.Fatalf("failed to power on source server %s: %s", src.Name, err)
			}
			log.Printf("Waiting for %s to power on (status ID: %s) ...", src.Name, reqID)
			if _, err = client.AwaitCompletion(reqID); err != nil {
				exit.Fatalf("failed to await completion of %s: %s", reqID, err)
			}
			// When the server is being powered on, it can take up to 5 minutes until
			// the backend is able to clone it; it requires the server to be fully booted.
			maxAttempts = 9
			time.Sleep(1 * time.Minute)
		}

		// We need the credentials, too
		log.Printf("Obtaining %s credentials ...", src.Name)
		credentials, err := client.GetServerCredentials(src.Name)
		if err != nil {
			exit.Fatalf("failed to obtain the credentials of server %q: %s", src.Name, err)
		}

		req := clcv2.CreateServerReq{
			Name:                 cloneFlags.seed,
			Cpu:                  src.Details.Cpu,
			MemoryGB:             src.Details.MemoryMb >> 10,
			GroupId:              src.GroupId,
			SourceServerId:       src.Name,
			PrimaryDns:           cloneFlags.primDNS,
			SecondaryDns:         cloneFlags.secDNS,
			Password:             credentials.Password,
			SourceServerPassword: credentials.Password,

			Type: src.Type,
		}

		if cloneFlags.seed == "" {
			if l := len(src.Name); l > 10 { // use same naming as original by default
				req.Name = strings.TrimRight(src.Name[7:l-1], "0")
			} else {
				req.Name = "CLONE"
			}
		}

		if cloneFlags.numCpu != 0 {
			req.Cpu = int(cloneFlags.numCpu)
		}
		if cloneFlags.memGB != 0 {
			req.MemoryGB = int(cloneFlags.memGB)
		}

		if cloneFlags.desc != "" {
			req.Description = cloneFlags.desc
		} else if src.Description == "" {
			req.Description = fmt.Sprintf("Clone of %s", src.Name)
		} else {
			req.Description = fmt.Sprintf("%s (cloned from %s)", src.Description, src.Name)
		}

		if cloneFlags.extraDrv != 0 {
			req.AdditionalDisks = append(req.AdditionalDisks,
				clcv2.ServerAdditionalDisk{SizeGB: cloneFlags.extraDrv, Type: "raw"})
		}
		if cloneFlags.ttl != 0 { /* Date/time that the server should be deleted. */
			req.Ttl = new(time.Time)
			*req.Ttl = time.Now().Add(cloneFlags.ttl)
		}

		/* hwGroup may be hex uuid or group name */
		if dest != "" {
			req.GroupId = dest

			if _, err := hex.DecodeString(dest); err != nil {
				log.Printf("Resolving ID of Hardware Group %q in %s ...", dest, src.LocationId)

				if group, err := client.GetGroupByName(dest, src.LocationId); err != nil {
					exit.Fatalf("failed to resolve group name %q: %s", dest, err)
				} else if group == nil {
					exit.Errorf("No group named %q was found in %s", dest, src.LocationId)
				} else {
					req.GroupId = group.Id
				}
			}
		}

		/* net is supposed to be a (hex) ID, but allow network names, too */
		if cloneFlags.net == "" {
			log.Printf("Determining network ID used by %s ...", src.Name)

			nets, err := client.GetServerNets(src)
			if err != nil {
				exit.Fatalf("failed to query networks of %s: %s", src.Name, err)
			}

			if len(nets) == 0 {
				// No network information found for the server, even though it has an IP.
				// This can happen when the server is owned by a sub-account, and uses a
				// network that is owned by the parent account. In this case, the sub-account
				// is prevented from querying details of the parent account, due to insufficient
				// permission.
				if parentAccount := client.RegisteredAccountAlias(); client.AccountAlias != parentAccount {
					var savedAlias = client.AccountAlias

					log.Printf("Network ID not visible under %q account - trying %q instead ...", savedAlias, parentAccount)
					client.AccountAlias = parentAccount
					if nets, err = client.GetServerNets(src); err != nil {
						exit.Fatalf("failed to query networks of %s using %q account: %s", src.Name, parentAccount, err)
					}
					// Restore Account Alias for remainder of program
					client.AccountAlias = savedAlias
				}
				if len(nets) == 0 {
					log.Printf("Unable to determine Network ID - querying %s deployable networks ...", src.LocationId)
					capa, err := client.GetDeploymentCapabilities(src.LocationId)
					if err != nil {
						exit.Fatalf("failed to determine %s Deployment Capabilities: %s", src.LocationId, err)
					}
					fmt.Println("Please specify the network ID for the clone manually via -net, using this information:")
					table := tablewriter.NewWriter(os.Stdout)
					table.SetAutoFormatHeaders(false)
					table.SetAlignment(tablewriter.ALIGN_LEFT)
					table.SetAutoWrapText(false)

					table.SetHeader([]string{"Name", "Type", "Account", "Network ID"})
					for _, net := range capa.DeployableNetworks {
						table.Append([]string{net.Name, net.Type, net.AccountID, net.NetworkId})
					}

					table.Render()
					os.Exit(0)
				}
			}

			if len(nets) != 1 {
				// FIXME: print server networks
				exit.Errorf("please specify which network to use (%s uses %d)", src.Name, len(nets))
			} else {
				log.Printf("Using %s network %s, with gateway %s", nets[0].Type, nets[0].Cidr, nets[0].Gateway)
				req.NetworkId = nets[0].Id
			}
		} else if _, err := hex.DecodeString(cloneFlags.net); err != nil { // not a HEX ID, treat as group name
			log.Printf("Resolving network ID of %q in %s ...", cloneFlags.net, src.LocationId)

			if netw, err := client.GetNetworkIdByName(cloneFlags.net, src.LocationId); err != nil {
				exit.Fatalf("failed to resolve network name %q: %s", cloneFlags.net, err)
			} else if netw == nil {
				exit.Fatalf("unable to resolve network name %q in %s - maybe use hex ID?", cloneFlags.net, src.LocationId)
			} else {
				req.NetworkId = netw.Id
			}
		} else { // HEX ID, use directly
			req.NetworkId = cloneFlags.net
		}

		for i := 1; ; i++ {
			url, reqID, err = client.CreateServer(&req)
			if err == nil || i == maxAttempts || strings.Index(err.Error(), "body.sourceServerId") > 0 {
				break
			}
			log.Printf("attempt %d/%d failed (%s) - retrying ...", i, maxAttempts, strings.TrimSpace(err.Error()))
			time.Sleep(1 * time.Minute)
		}
		if err != nil {
			exit.Fatalf("failed to create server: %s", err)
		}
		log.Printf("Cloning %s: %s", src.Name, reqID)

		status, err := client.PollStatusFn(reqID, intvl, func(s clcv2.QueueStatus) {
			log.Printf("Cloning %s => %s/%s..%s..: %s", source, dest, src.LocationId, req.Name, s)
		})
		if err != nil {
			exit.Fatalf("failed to poll %s status: %s", reqID, err)
		}

		server, err := client.GetServerByURI(url)
		if err != nil {
			log.Fatalf("failed to query server details at %s: %s", url, err)
		} else if status == clcv2.Failed {
			exit.Fatalf("failed to clone %s (will show up as 'under construction')", server.Name)
		}
		log.Printf("New server name: %s\n", server.Name)
		log.Printf("Server Password: \"%s\"\n", credentials.Password)
	},
}
