package main

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the application backend bound to the frontend. It owns the media
// Library and exposes methods for loading tracks and controlling the window.
type App struct {
	ctx       context.Context
	lib       *Library
	mediaBase string

	// progress holds per-track playback positions (track id -> seconds), loaded
	// at startup and updated as the user listens so playback can resume.
	progressMu sync.Mutex
	progress   map[string]float64
}

// NewApp creates a new App application struct. mediaBase is the base URL of the
// loopback media server (e.g. "http://127.0.0.1:52341").
func NewApp(lib *Library, mediaBase string) *App {
	return &App{lib: lib, mediaBase: mediaBase, progress: loadProgress()}
}

// MediaBaseURL exposes the media server's base URL so the frontend can build
// absolute URLs for audio, cover art and lyrics.
func (a *App) MediaBaseURL() string {
	return a.mediaBase
}

// startup is called when the app starts. The context is saved so we can call
// the runtime methods, and we register the native file-drop handler so tracks
// dragged onto the window get imported.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	runtime.OnFileDrop(ctx, func(_, _ int, paths []string) {
		tracks := a.AddPaths(paths)
		if len(tracks) > 0 {
			runtime.EventsEmit(ctx, "tracks:dropped", tracks)
		}
	})
}

// GetInitialTracks restores the user's music library on launch. It first tries
// the persisted library (paths the user imported previously); if that is empty
// or all its files are gone, it falls back to the bundled "audios" sample so a
// fresh install still has something to play. The resulting library is persisted
// so stale entries get pruned.
func (a *App) GetInitialTracks() []*Track {
	var tracks []*Track

	if paths := loadPersistedPaths(); len(paths) > 0 {
		tracks = a.lib.scanFiles(paths)
	}

	if len(tracks) == 0 {
		if dir := a.defaultAudioDir(); dir != "" {
			tracks = a.lib.scanDirectory(dir)
		}
	}

	a.persist()
	if tracks == nil {
		return []*Track{}
	}
	return tracks
}

// persist writes the current library paths to disk, ignoring IO errors (the app
// stays usable even if the config directory is not writable).
func (a *App) persist() {
	_ = persistPaths(a.lib.orderedPaths())
}

// defaultAudioDir returns the first existing "audios" directory among a set of
// candidate locations, or "" if none exist. Used only as a first-run sample.
func (a *App) defaultAudioDir() string {
	var candidates []string
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), "audios"))
	}
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(wd, "audios"))
	}
	candidates = append(candidates, "audios")

	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			return c
		}
	}
	return ""
}

// OpenFolder shows a directory picker and returns every supported audio track
// found inside (recursively).
func (a *App) OpenFolder() ([]*Track, error) {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择音乐文件夹",
	})
	if err != nil {
		return nil, err
	}
	if dir == "" {
		return []*Track{}, nil
	}
	tracks := a.lib.scanDirectory(dir)
	a.persist()
	if tracks == nil {
		return []*Track{}, nil
	}
	return tracks, nil
}

// OpenFiles shows a multi-file picker filtered to supported audio formats.
func (a *App) OpenFiles() ([]*Track, error) {
	paths, err := runtime.OpenMultipleFilesDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择音乐文件",
		Filters: []runtime.FileFilter{
			{DisplayName: "音频文件 (*.mp3;*.flac;*.wav;*.m4a;*.ogg)", Pattern: "*.mp3;*.flac;*.wav;*.m4a;*.ogg"},
			{DisplayName: "所有文件 (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil {
		return nil, err
	}
	tracks := a.lib.scanFiles(paths)
	a.persist()
	if tracks == nil {
		return []*Track{}, nil
	}
	return tracks, nil
}

// AddPaths ingests a mix of files and directories (used by drag-and-drop),
// registering every supported track found and persisting the library.
func (a *App) AddPaths(paths []string) []*Track {
	var out []*Track
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		if info.IsDir() {
			out = append(out, a.lib.scanDirectory(p)...)
		} else {
			out = append(out, a.lib.scanFiles([]string{p})...)
		}
	}
	a.persist()
	if out == nil {
		return []*Track{}
	}
	return out
}

// RemoveTrack drops a single track from the library and persists the change.
func (a *App) RemoveTrack(id string) {
	a.lib.remove(id)
	a.persist()
	a.ClearProgress(id)
}

// ClearLibrary empties the whole library and persists the change.
func (a *App) ClearLibrary() {
	a.lib.clear()
	a.persist()
	a.progressMu.Lock()
	a.progress = map[string]float64{}
	a.progressMu.Unlock()
	_ = saveProgressFile(map[string]float64{})
}

// ---- Playback progress ----

// GetProgress returns the saved playback position (seconds) for every track id
// that has one, so the frontend can resume where the user left off.
func (a *App) GetProgress() map[string]float64 {
	a.progressMu.Lock()
	defer a.progressMu.Unlock()
	return a.cloneProgressLocked()
}

// SaveProgress records the playback position (seconds) for a track and writes
// it to disk so it survives restarts.
func (a *App) SaveProgress(id string, seconds float64) {
	if id == "" || seconds <= 0 {
		return
	}
	a.progressMu.Lock()
	a.progress[id] = seconds
	snapshot := a.cloneProgressLocked()
	a.progressMu.Unlock()
	_ = saveProgressFile(snapshot)
}

// ClearProgress forgets a track's saved position (e.g. once it finishes).
func (a *App) ClearProgress(id string) {
	a.progressMu.Lock()
	if _, ok := a.progress[id]; !ok {
		a.progressMu.Unlock()
		return
	}
	delete(a.progress, id)
	snapshot := a.cloneProgressLocked()
	a.progressMu.Unlock()
	_ = saveProgressFile(snapshot)
}

// cloneProgressLocked returns a copy of the progress map. Callers must hold
// progressMu.
func (a *App) cloneProgressLocked() map[string]float64 {
	out := make(map[string]float64, len(a.progress))
	for k, v := range a.progress {
		out[k] = v
	}
	return out
}

// ---- Window controls (frameless window) ----

func (a *App) WindowMinimise() {
	runtime.WindowMinimise(a.ctx)
}

func (a *App) WindowToggleMaximise() {
	runtime.WindowToggleMaximise(a.ctx)
}

func (a *App) Quit() {
	runtime.Quit(a.ctx)
}
