package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/gado-ships-it/news-cli/internal/source"
)

// Store is the on-disk representation of the user's local source list.
// Seeded sources are merged in at load time; the on-disk file only holds
// user additions/overrides so upstream updates flow through cleanly.
type Store struct {
	Sources map[string]source.Source `json:"sources"`
}

// Dir returns the config dir, creating it if needed.
func Dir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, "news-cli")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

// Path returns the sources.json path inside the config dir.
func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "sources.json"), nil
}

// Load reads the on-disk store; missing file returns an empty store.
func Load() (*Store, error) {
	p, err := Path()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if errors.Is(err, os.ErrNotExist) {
		return &Store{Sources: map[string]source.Source{}}, nil
	}
	if err != nil {
		return nil, err
	}
	var s Store
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse %s: %w", p, err)
	}
	if s.Sources == nil {
		s.Sources = map[string]source.Source{}
	}
	return &s, nil
}

// Save writes the store atomically.
func (s *Store) Save() error {
	p, err := Path()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

// Add or replace a source in the store and stamp LearnedAt.
func (s *Store) Add(src source.Source) {
	if src.LearnedAt == nil {
		now := time.Now().UTC()
		src.LearnedAt = &now
	}
	s.Sources[src.Name] = src
}

// Remove a source from the store. Returns true if it existed.
func (s *Store) Remove(name string) bool {
	if _, ok := s.Sources[name]; !ok {
		return false
	}
	delete(s.Sources, name)
	return true
}

// Merged returns seed sources overlaid with user store (user wins on name collision).
// Result is sorted by Name for stable output.
func Merged(seed []source.Source, store *Store) []source.Source {
	byName := map[string]source.Source{}
	for _, s := range seed {
		byName[s.Name] = s
	}
	for k, v := range store.Sources {
		byName[k] = v
	}
	out := make([]source.Source, 0, len(byName))
	for _, v := range byName {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}
