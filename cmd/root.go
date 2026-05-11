package cmd

import (
	"github.com/spf13/cobra"
)

// Global flags shared across every subcommand.
var (
	flagMD   bool
	flagJSON bool
)

var rootCmd = &cobra.Command{
	Use:   "news",
	Short: "Read public news frontpages from the terminal.",
	Long: `news-cli fetches headlines, deks and lede images from publicly
available news outlet frontpages.

Every output prominently references the original source: results carry the
source name, the source homepage URL, and a direct link to the article on
the publisher's own site. The tool is designed to send traffic back to
publishers, not to replace reading them.

Default output is terse, terminal-friendly text. Use --md for Markdown or
--json for structured output that an LLM (or any downstream script) can
consume.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&flagMD, "md", false, "render output as Markdown")
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "render output as JSON")

	rootCmd.AddCommand(fetchCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(learnCmd)
	rootCmd.AddCommand(submitCmd)
	rootCmd.AddCommand(sourcesCmd)
}
