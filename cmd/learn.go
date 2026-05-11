package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/gado-ships-it/news-cli/internal/config"
	"github.com/gado-ships-it/news-cli/internal/learn"
	"github.com/gado-ships-it/news-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	learnName    string
	learnTitle   string
	learnYes     bool
	learnTimeout time.Duration
)

var learnCmd = &cobra.Command{
	Use:   "learn <homepage-url>",
	Short: "Derive an extractor for a new news source.",
	Long: `learn tries, in order:

  1. RSS / Atom autodiscovery (link rel="alternate") and common feed paths
  2. schema.org NewsArticle blocks in <script type="application/ld+json">
  3. LLM-derived CSS selectors against the homepage HTML (requires
     ANTHROPIC_API_KEY)

When a method succeeds the derived source is saved to your local config and
you are prompted to submit it back to the shared gado-ships-it/news-cli source list
as a pull request (skip the prompt with --yes).`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		homepage := args[0]
		ctx, cancel := context.WithTimeout(cmd.Context(), learnTimeout)
		defer cancel()

		// LLM is optional — only needed for the CSS fallback. We construct
		// it eagerly so the user gets a clear error before scraping the page
		// rather than after.
		llm, llmErr := learn.NewBackendFromEnv()
		if llmErr != nil {
			fmt.Fprintf(os.Stderr, "note: %s — CSS fallback disabled\n", llmErr)
		}

		res, err := learn.Learn(ctx, homepage, llm)
		if err != nil {
			return err
		}
		if learnName != "" {
			res.Source.Name = learnName
		}
		if learnTitle != "" {
			res.Source.Title = learnTitle
		}

		// Persist to the local store.
		store, err := config.Load()
		if err != nil {
			return err
		}
		store.Add(res.Source)
		if err := store.Save(); err != nil {
			return err
		}

		// Show the result.
		format := output.FromFlags(flagMD, flagJSON)
		switch format {
		case output.FormatJSON:
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			if err := enc.Encode(res); err != nil {
				return err
			}
		case output.FormatMD:
			fmt.Printf("# Learned `%s`\n\n", res.Source.Name)
			fmt.Printf("- Title: %s\n- Homepage: %s\n- Method: **%s**\n- Items found: **%d**\n\n", res.Source.Title, res.Source.Homepage, res.Method, res.ItemsFound)
			fmt.Println("Sample headlines:")
			for _, it := range res.Sample {
				fmt.Printf("- [%s](%s)\n", it.Headline, it.URL)
			}
			fmt.Println()
		default:
			fmt.Printf("learned %q via %s — %d items found\n", res.Source.Name, res.Method, res.ItemsFound)
			fmt.Printf("  title:    %s\n", res.Source.Title)
			fmt.Printf("  homepage: %s\n", res.Source.Homepage)
			for _, it := range res.Sample {
				fmt.Printf("  • %s\n", it.Headline)
			}
		}

		// PR-back prompt unless caller suppressed it.
		if !learnYes {
			fmt.Fprintln(os.Stderr)
			fmt.Fprintln(os.Stderr, "Share this source with the community by submitting a PR to gado-ships-it/news-cli?")
			fmt.Fprintln(os.Stderr, "Run: news submit "+res.Source.Name)
		}
		return nil
	},
}

func init() {
	learnCmd.Flags().StringVar(&learnName, "name", "", "override the auto-derived short name")
	learnCmd.Flags().StringVar(&learnTitle, "title", "", "override the auto-derived display title")
	learnCmd.Flags().BoolVar(&learnYes, "yes", false, "skip the PR-back prompt")
	learnCmd.Flags().DurationVar(&learnTimeout, "timeout", 90*time.Second, "overall learn timeout")
}
