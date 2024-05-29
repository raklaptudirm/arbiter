package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	arbiter "laptudirm.com/x/arbiter/pkg/manager"
)

func Remove() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove engine",
		Short: "Uninstall the given engine",
		Args:  cobra.ExactArgs(1),

		RunE: func(cmd *cobra.Command, args []string) error {
			source, tag, has_tag := strings.Cut(args[0], "@")

			engine, err := arbiter.NewEngine(source)
			if err != nil {
				return err
			}

			if has_tag {
				version, err := engine.ResolveVersion(tag)
				if err != nil {
					return err
				}
				if !arbiter.Downloaded(engine, version) {
					fmt.Printf("\nEngine \x1b[32m%s %s\x1b[0m is not installed.\n", engine.Name, version.Name)
					return nil
				}

				fmt.Printf("\x1b[32mUninstalling Engine:\x1b[0m %s %s\n\n", engine.Name, tag)
				arbiter.Engines.RemoveVersion(engine, version.Name)
				os.Remove(arbiter.VersionBinary(engine, version.Name))
				return nil
			}

			fmt.Printf("\x1b[32mUninstalling Engine:\x1b[0m %s\n\n", engine.Name)
			os.Remove(arbiter.Binary(engine))
			for _, version := range arbiter.Engines[engine.Name].Versions {
				os.Remove(arbiter.VersionBinary(engine, version))
			}
			arbiter.Engines.RemoveEngine(engine)

			return nil
		},
	}

	cmd.Flags().BoolP("force", "f", false, "Force a re-installation of the engine")
	cmd.Flags().BoolP("no-main", "n", false, "Don't replace the engine with the new version")

	return cmd
}
