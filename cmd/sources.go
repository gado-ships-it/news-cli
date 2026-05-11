package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/gado-ships-it/news-cli/internal/config"
	"github.com/spf13/cobra"
)

var sourcesCmd = &cobra.Command{
	Use:   "sources",
	Short: "Manage local source overrides.",
}

var sourcesRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a source from the local store. Seeded sources are unaffected.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := config.Load()
		if err != nil {
			return err
		}
		if !store.Remove(args[0]) {
			return fmt.Errorf("source %q not found in local store", args[0])
		}
		return store.Save()
	},
}

var sourcesPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Print the path to the local sources.json store.",
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := config.Path()
		if err != nil {
			return err
		}
		fmt.Println(p)
		return nil
	},
}

var sourcesShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show the full JSON definition of one configured source.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := config.Load()
		if err != nil {
			return err
		}
		// Prefer the local override; fall back to seed.
		if s, ok := store.Sources[args[0]]; ok {
			return json.NewEncoder(os.Stdout).Encode(s)
		}
		// scan seed
		for _, s := range seedAllSources() {
			if s.Name == args[0] {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(s)
			}
		}
		return fmt.Errorf("source %q not found", args[0])
	},
}

func init() {
	sourcesCmd.AddCommand(sourcesRemoveCmd, sourcesPathCmd, sourcesShowCmd)
}
