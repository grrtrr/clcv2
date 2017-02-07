package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
)

// location of the bash auto-completion file
var autoCompleteFile string

// inspired by hugo autocomplete
var autoComplete = &cobra.Command{
	Use:    "bash-completion",
	Hidden: true,
	Short:  fmt.Sprintf("Generate  autocompletion script for %s", path.Base(os.Args[0])),
	Long:   `Generates a shell autocompletion script, to be sourced via e.g. /etc/bash_completion.d`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := cmd.Root().GenBashCompletionFile(autoCompleteFile); err != nil {
			return err
		}

		fmt.Printf("Bash completion file for saved to %q.\n", autoCompleteFile)
		return nil
	},
}

func init() {
	autoComplete.PersistentFlags().StringVarP(&autoCompleteFile, "filename", "f",
		fmt.Sprintf("/etc/bash_completion.d/%s.sh", path.Base(os.Args[0])), "Path to bash autocompletion file")
	Root.AddCommand(autoComplete)
}
