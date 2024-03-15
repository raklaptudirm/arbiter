package cmd

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"

	"laptudirm.com/x/arbiter/pkg/manager"
)

// arbiter install
func Install() *cobra.Command {
	return &cobra.Command{
		Use:   "install { engine owner/engine git-url }",
		Short: "Install the given Game Player",
		Args:  cobra.ExactArgs(1),

		Long: heredoc.Doc(`install installs the given engine into arbiter so that it
			can be used globally and in arbiter without configuration
			and sourcing the engine every time it is used.

			The formats supported for the engine name are <name>,
			<owner>/<name> (for engines on github), or a full <url> to
			a git repository. The <name> format is only supported for
			the engines arbiter is configured by default for.

			To use the installed engines from the command line, you will
			need to add the directory ~/arbiter to your path variable.`),

		RunE: func(cmd *cobra.Command, args []string) error {
			repo, err := manager.NewBareRepository(args[0])
			if err != nil {
				return err
			}

			return repo.InstallEngine()
		},
	}
}
