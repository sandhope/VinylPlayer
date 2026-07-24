package main

import (
	"crypto/sha1"
	"encoding/hex"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/dhowden/tag"
)

// Track describes a single audio file exposed to the frontend.
type Track struct {
	ID       string `json:"id"`
	Path     string `json:"path"`
	Title    string `json:"title"`
	Artist   string `json:"artist"`
	Album    string `json:"album"`
	Format   string `json:"format"`
	URL      string `json:"url"`
	CoverURL string `json:"coverUrl"`
	LyricURL string `json:"lyricUrl"`

	// cover data kept in-memory so the media server can serve it without
	// re-parsing the file. Unexported, so never serialized to the frontend.
	coverData []byte
	coverMime string
	// lyricFile is the resolved on-disk path of the sibling .lrc file (if any),
	// so the media server can read it without re-deriving the name.
	lyricFile string
}

// supportedExts lists the audio containers the player understands. WebView2
// (Chromium) decodes all of these natively, including FLAC.
var supportedExts = map[string]string{
	".mp3":  "MP3",
	".flac": "FLAC",
	".wav":  "WAV",
	".m4a":  "M4A",
	".ogg":  "OGG",
}

// Library keeps the registry of known tracks so the media server can resolve
// an opaque id back to an on-disk path. It also keeps an ordered list of the
// absolute paths in the library, which is what gets persisted across restarts.
type Library struct {
	mu     sync.RWMutex
	tracks map[string]*Track
	order  []string // ordered, de-duplicated absolute paths
}

func NewLibrary() *Library {
	return &Library{tracks: make(map[string]*Track)}
}

func (l *Library) get(id string) (*Track, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	t, ok := l.tracks[id]
	return t, ok
}

func (l *Library) register(t *Track) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, exists := l.tracks[t.ID]; !exists {
		l.order = append(l.order, t.Path)
	}
	l.tracks[t.ID] = t
}

// orderedPaths returns the current library paths in insertion order.
func (l *Library) orderedPaths() []string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := make([]string, len(l.order))
	copy(out, l.order)
	return out
}

// remove deletes a track (by id) from the registry and drops its path from the
// ordered list, so the change is reflected in the persisted library.
func (l *Library) remove(id string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	t, ok := l.tracks[id]
	if !ok {
		return
	}
	delete(l.tracks, id)
	for i, p := range l.order {
		if p == t.Path {
			l.order = append(l.order[:i], l.order[i+1:]...)
			break
		}
	}
}

// clear empties the whole library.
func (l *Library) clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.tracks = make(map[string]*Track)
	l.order = nil
}

// idForPath produces a stable id from the absolute path.
func idForPath(path string) string {
	sum := sha1.Sum([]byte(strings.ToLower(path)))
	return hex.EncodeToString(sum[:])[:16]
}

// scanFiles builds Track entries for the given absolute file paths, reading
// embedded metadata where available and registering them for serving.
func (l *Library) scanFiles(paths []string) []*Track {
	var out []*Track
	for _, p := range paths {
		abs, err := filepath.Abs(p)
		if err != nil {
			continue
		}
		ext := strings.ToLower(filepath.Ext(abs))
		format, ok := supportedExts[ext]
		if !ok {
			continue
		}
		// Skip paths that no longer exist (e.g. stale persisted entries).
		if info, statErr := os.Stat(abs); statErr != nil || info.IsDir() {
			continue
		}
		t := l.buildTrack(abs, format)
		l.register(t)
		out = append(out, t)
	}
	return out
}

// scanDirectory walks a directory (non-recursive by default is not enough for a
// music player, so we walk recursively) and returns discovered tracks sorted by
// title.
func (l *Library) scanDirectory(dir string) []*Track {
	var paths []string
	_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if _, ok := supportedExts[ext]; ok {
			paths = append(paths, path)
		}
		return nil
	})
	tracks := l.scanFiles(paths)
	sort.SliceStable(tracks, func(i, j int) bool {
		return strings.ToLower(tracks[i].Title) < strings.ToLower(tracks[j].Title)
	})
	return tracks
}

// buildTrack extracts metadata for a single file. Missing tags gracefully fall
// back to the file name so the UI always has something sensible to display.
func (l *Library) buildTrack(abs, format string) *Track {
	id := idForPath(abs)
	base := strings.TrimSuffix(filepath.Base(abs), filepath.Ext(abs))

	t := &Track{
		ID:     id,
		Path:   abs,
		Title:  base,
		Artist: "未知艺术家",
		Album:  "",
		Format: format,
		URL:    "/media/" + id,
	}

	if f, err := os.Open(abs); err == nil {
		defer f.Close()
		if m, err := tag.ReadFrom(f); err == nil {
			if v := strings.TrimSpace(m.Title()); v != "" {
				t.Title = v
			}
			if v := strings.TrimSpace(m.Artist()); v != "" {
				t.Artist = v
			}
			if v := strings.TrimSpace(m.Album()); v != "" {
				t.Album = v
			}
			if pic := m.Picture(); pic != nil && len(pic.Data) > 0 {
				t.coverData = pic.Data
				t.coverMime = pic.MIMEType
				if t.coverMime == "" {
					t.coverMime = "image/jpeg"
				}
				t.CoverURL = "/cover/" + id
			}
		}
	}

	// Look for a sibling .lrc lyric file. Real-world files often carry an
	// artist suffix (e.g. "水中花.mp3" alongside "水中花 - 谭咏麟.lrc"), so match
	// more than the exact base name.
	if lrc := findLyricFile(abs, t.Artist); lrc != "" {
		t.lyricFile = lrc
		t.LyricURL = "/lyric/" + id
	}

	return t
}

// findLyricFile locates a sibling .lrc lyric file for the given audio path,
// trying in order: an exact "<base>.lrc"; a "<base> - <artist>.lrc" variant
// (a common naming convention); and finally any .lrc in the same directory
// whose name starts with the audio base name followed by a separator. Returns
// an empty string when nothing matches.
func findLyricFile(audioPath, artist string) string {
	dir := filepath.Dir(audioPath)
	base := strings.TrimSuffix(filepath.Base(audioPath), filepath.Ext(audioPath))

	exact := filepath.Join(dir, base+".lrc")
	if fi, err := os.Stat(exact); err == nil && !fi.IsDir() {
		return exact
	}
	if artist != "" {
		withArtist := filepath.Join(dir, base+" - "+artist+".lrc")
		if fi, err := os.Stat(withArtist); err == nil && !fi.IsDir() {
			return withArtist
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	lowBase := strings.ToLower(base)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.EqualFold(filepath.Ext(name), ".lrc") {
			continue
		}
		low := strings.ToLower(name)
		if !strings.HasPrefix(low, lowBase) {
			continue
		}
		// Require a separator after the base name so "水" doesn't match
		// "水中花 - ...".
		rest := low[len(lowBase):]
		if rest == ".lrc" || strings.HasPrefix(rest, " ") ||
			strings.HasPrefix(rest, "-") || strings.HasPrefix(rest, "_") ||
			strings.HasPrefix(rest, ".") {
			return filepath.Join(dir, name)
		}
	}
	return ""
}
