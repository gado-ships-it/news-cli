package cmd

import (
	"github.com/gado-ships-it/news-cli/internal/seed"
	"github.com/gado-ships-it/news-cli/internal/source"
)

// seedAllSources is a tiny indirection so the sources subcommand can read
// the bundled list without a direct import in every file.
func seedAllSources() []source.Source { return seed.Sources() }
