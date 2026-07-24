package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// persistedLibrary is the on-disk shape of the user's music library. We store
// absolute file paths (not parsed metadata) so the library survives restarts
// and metadata is always re-read fresh from the files.
type persistedLibrary struct {
	Version int      `json:"version"`
	Tracks  []string `json:"tracks"`
}

// libraryFilePath returns the location of the persisted library file inside the
// per-user config directory, e.g. %AppData%/VinylPlayer/library.json on Windows.
func libraryFilePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "VinylPlayer", "library.json"), nil
}

// loadPersistedPaths reads the saved track paths. A missing or unreadable file
// simply yields an empty list (first run).
func loadPersistedPaths() []string {
	path, err := libraryFilePath()
	if err != nil {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var lib persistedLibrary
	if err := json.Unmarshal(data, &lib); err != nil {
		return nil
	}
	return lib.Tracks
}

// persistPaths writes the current library paths to disk, creating the config
// directory if needed.
func persistPaths(paths []string) error {
	path, err := libraryFilePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(persistedLibrary{Version: 1, Tracks: paths}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
