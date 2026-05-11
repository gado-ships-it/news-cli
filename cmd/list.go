package cmd

import (
	"os"

	"github.com/gado-ships-it/news-cli/internal/config"
	"github.com/gado-ships-it/news-cli/internal/output"
	"github.com/gado-ships-it/news-cli/internal/seed"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured sources (seeded + locally learned).",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := config.Load()
		if err != nil {
			return err
		}
		all := config.Merged(seed.Sources(), store)
		return output.Sources(os.Stdout, all, output.FromFlags(flagMD, flagJSON))
	},
}
