package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/gado-ships-it/news-cli/internal/config"
	"github.com/gado-ships-it/news-cli/internal/output"
	"github.com/spf13/cobra"
)

var submitCmd = &cobra.Command{
	Use:   "submit <name>",
	Short: "Open a PR to gado-ships-it/news-cli adding a learned source.",
	Long: `submit packages a locally-learned source as a contribution to the shared
gado-ships-it/news-cli source list. If the GitHub CLI ("gh") is installed and
authenticated, a pull request is opened automatically; otherwise the
encoded source is printed along with manual instructions.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		store, err := config.Load()
		if err != nil {
			return err
		}
		src, ok := store.Sources[name]
		if !ok {
			return fmt.Errorf("source %q is not in your local store — only learned sources can be submitted", name)
		}

		// Pretty-print the source as the canonical JSON we want appended
		// to the shared list.
		data, err := json.MarshalIndent(src, "", "  ")
		if err != nil {
			return err
		}

		format := output.FromFlags(flagMD, flagJSON)
		switch format {
		case output.FormatJSON:
			fmt.Println(string(data))
		case output.FormatMD:
			fmt.Println("```json")
			fmt.Println(string(data))
			fmt.Println("```")
		}

		// Best-effort PR via gh.
		if _, err := exec.LookPath("gh"); err != nil {
			fmt.Fprintln(os.Stderr)
			fmt.Fprintln(os.Stderr, "gh CLI not found — submit manually:")
			fmt.Fprintln(os.Stderr, "  1. Fork https://github.com/gado-ships-it/news-cli")
			fmt.Fprintln(os.Stderr, "  2. Append the JSON above to internal/seed/sources.go")
			fmt.Fprintln(os.Stderr, "  3. Open a PR titled: \"add source: "+src.Title+"\"")
			return nil
		}

		body := fmt.Sprintf("Adds %s (%s) — learned via `news learn` (extractor type: `%s`).\n\nSource definition:\n```json\n%s\n```\n", src.Title, src.Homepage, src.Extractor.Type, string(data))
		title := "add source: " + src.Title

		fmt.Fprintf(os.Stderr, "Opening PR via gh: %q\n", title)
		c := exec.Command("gh", "pr", "create",
			"--repo", "gado-ships-it/news-cli",
			"--title", title,
			"--body", body,
			"--web", // open in browser so the user reviews before pushing
		)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			return fmt.Errorf("gh pr create: %w (you can still submit manually)", err)
		}
		return nil
	},
}
