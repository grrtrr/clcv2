package cmd

/*
 * Setting/changing Server Attributes
 */
import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
	garbler "github.com/michaelbironneau/garbler/lib"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	Root.AddCommand(&cobra.Command{
		Use:     "cpu  <server> <numCPU>",
		Aliases: []string{"cores"},
		Short:   "Set server #CPU",
		Long:    "Sets the number of CPUs on @serverCPU to @numCPU",
		Example: "cpu WA1GRRT-W12-29 4",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return errors.Errorf("Need a <server> and a <#CPUs> argument")
			} else if _, err := strconv.ParseUint(args[1], 10, 8); err != nil {
				return errors.Errorf("Invalid numCPU value %q", args[1])
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Setting %s #CPUs to %s ...\n", args[0], args[1])
			if reqID, err := client.ServerSetCpus(args[0], args[1]); err != nil {
				exit.Fatalf("failed to change the number of CPUs on %q: %s", args[0], err)
			} else {
				log.Printf("%s changing number-of-CPUs: %s", args[0], reqID)

				client.PollStatusFn(reqID, intvl, func(s clcv2.QueueStatus) {
					log.Printf("%s changing number-of-CPUs: %s", args[0], s)
				})
			}
		},
	})

	Root.AddCommand(&cobra.Command{
		Use:     "memory  <server> <memoryGB>",
		Aliases: []string{"mem", "ram"},
		Short:   "Set server memory",
		Long:    "Sets the memory of @server to size @memoryGB",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return errors.Errorf("Need a <server> and a <memoryGB> argument")
			} else if _, err := strconv.ParseUint(args[1], 10, 32); err != nil {
				return errors.Errorf("Invalid memoryGB value %q", args[1])
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Setting %s memory to %s GB ...\n", args[0], args[1])
			if reqID, err := client.ServerSetMemory(args[0], args[1]); err != nil {
				exit.Fatalf("failed to change the amount of memory on %q: %s", args[0], err)
			} else {
				log.Printf("%s changing memory size: %s", args[0], reqID)

				client.PollStatusFn(reqID, intvl, func(s clcv2.QueueStatus) {
					log.Printf("%s changing memory size: %s", args[0], s)
				})
			}
		},
	})

	Root.AddCommand(&cobra.Command{
		Use:     "desc  <server>",
		Aliases: []string{"description"},
		Short:   "Power-off server(s)",
		Long:    "Do a forceful power-off of server(s) (as opposed to a soft OS-level shutdown)",
		PreRunE: checkArgs(2, "Need a server name and a new description for the server"),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Setting %s description to to %q.\n", args[0], args[1])

			if err := client.ServerSetDescription(args[0], args[1]); err != nil {
				exit.Fatalf("failed to change the description of %q: %s", args[0], err)
			}
		},
	})

	Root.AddCommand(&cobra.Command{
		Use:     "password  <server>  [password]",
		Aliases: []string{"pass", "set-pass"},
		Short:   "Set or generate server password",
		Long:    "Sets a new password for @server if provided, or generates a paranoid 'garbler' password",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if l := len(args); l != 1 && l != 2 {
				return errors.Errorf("Need a server name and optionally a new password")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			var newPassword string

			log.Printf("Looking up existing password of %s", args[0])

			credentials, err := client.GetServerCredentials(args[0])
			if err != nil {
				exit.Fatalf("failed to obtain the credentials of %q: %s", args[0], err)
			}
			log.Printf("Existing %s password: %q", args[0], credentials.Password)

			if len(args) == 2 {
				newPassword = args[1]
			} else if newPassword, err = garbler.NewPassword(&garbler.Paranoid); err != nil {
				exit.Fatalf("failed to generate new 'garbler' password: %s", err)
			} else {
				// The 'Paranoid' mode in garbler more than satisfies CLC requirements.
				// However, the symbols may contain unsupported characters.
				newPassword = strings.Map(func(r rune) rune {
					if strings.Index(clcv2.InvalidPasswordCharacters, string(r)) > -1 {
						return '@'
					}
					return r
				}, newPassword)
				log.Printf("New paranoid 'garbler' password: %q", newPassword)
			}

			if newPassword == credentials.Password {
				log.Printf("%s password is already set to %q", args[1], newPassword)
				return
			}

			reqID, err := client.ServerChangePassword(args[0], credentials.Password, newPassword)
			if err != nil {
				exit.Fatalf("failed to change the password on %q: %s", args[0], err)
			} else {
				client.PollStatusFn(reqID, intvl, func(s clcv2.QueueStatus) {
					log.Printf("Updating %s password: %s", args[0], s)
				})
			}
		},
	})
}
