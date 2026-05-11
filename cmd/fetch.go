package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gado-ships-it/news-cli/internal/config"
	"github.com/gado-ships-it/news-cli/internal/fetcher"
	"github.com/gado-ships-it/news-cli/internal/output"
	"github.com/gado-ships-it/news-cli/internal/seed"
	"github.com/gado-ships-it/news-cli/internal/source"
	"github.com/spf13/cobra"
)

var (
	fetchMax         int
	fetchConcurrency int
	fetchTimeout     time.Duration
)

var fetchCmd = &cobra.Command{
	Use:   "fetch [name...]",
	Short: "Fetch headlines from one or more configured sources.",
	Long: `Fetch headlines from the named sources, or every configured source if
no names are given. Sources can be referenced by their short name
(e.g. "nzz", "nyt", "bbc") — see "news list" for the full set.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := config.Load()
		if err != nil {
			return err
		}
		all := config.Merged(seed.Sources(), store)

		var picked []source.Source
		if len(args) == 0 {
			picked = all
		} else {
			byName := map[string]source.Source{}
			for _, s := range all {
				byName[s.Name] = s
			}
			for _, name := range args {
				s, ok := byName[name]
				if !ok {
					return fmt.Errorf("unknown source %q — try `news list`", name)
				}
				picked = append(picked, s)
			}
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), fetchTimeout)
		defer cancel()
		fps := fetcher.FetchAll(ctx, picked, fetchConcurrency)

		return output.Frontpages(os.Stdout, fps, output.FromFlags(flagMD, flagJSON), fetchMax)
	},
}

func init() {
	fetchCmd.Flags().IntVarP(&fetchMax, "max", "n", 10, "max items per source (0 = unlimited)")
	fetchCmd.Flags().IntVar(&fetchConcurrency, "concurrency", 6, "parallel fetches across sources")
	fetchCmd.Flags().DurationVar(&fetchTimeout, "timeout", 30*time.Second, "overall fetch timeout")
}
