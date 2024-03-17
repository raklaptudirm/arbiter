package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"

	arbiter "laptudirm.com/x/arbiter/pkg/common"
	"laptudirm.com/x/arbiter/pkg/manager"
)

// arbiter install
func Install() *cobra.Command {
	cmd := &cobra.Command{
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
			source, tag, has_tag := strings.Cut(args[0], "@")
			if !has_tag {
				tag = "stable"
			}

			engine, err := manager.NewEngine(source)
			if err != nil {
				return err
			}

			fmt.Printf("\x1b[32mInstalling Player:\x1b[0m %s by %s\n\n", engine.Name, engine.Author)

			if err := engine.EfficientFetch(); err != nil {
				return err
			}

			version, err := engine.ResolveVersion(tag)
			if err != nil {
				return err
			}

			// Re-install the version only if it hasn't been installed previously.
			if !cmd.Flag("force").Changed && engine.Installed(version) {
				fmt.Printf("\nEngine \x1b[32m%s %s\x1b[0m is already installed.\n", engine.Name, version.Name)
			} else {
				if err := engine.InstallEngine(version); err != nil {
					return err
				}
			}

			// Hardlink the engine binary to the latest installation.
			_ = os.Remove(engine.Binary())
			_ = os.Link(engine.VersionBinary(version), engine.Binary())

			// Replace the main engine executable with the newly installed version.
			if !cmd.Flag("no-main").Changed {
				arbiter.Engines.SetMainVersion(engine.Name, version.Name)
			}
			return nil
		},
	}

	cmd.Flags().BoolP("force", "f", false, "Force a re-installation of the engine")
	cmd.Flags().BoolP("no-main", "n", false, "Don't replace the engine with the new version")

	return cmd
}
