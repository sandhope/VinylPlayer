package main

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// mediaServer serves audio streams, embedded cover art and lyric files for the
// tracks registered in the Library. It runs on its own loopback HTTP port so
// requests never collide with the frontend (in `wails dev` the Vite dev server
// otherwise answers unknown paths with index.html, which breaks media loading).
type mediaServer struct {
	lib *Library
}

func newMediaServer(lib *Library) *mediaServer {
	return &mediaServer{lib: lib}
}

// explicitMime maps audio extensions to a content type. We set this ourselves
// rather than relying on mime.TypeByExtension, which on Windows reads from the
// registry and can return empty/incorrect values (breaking playback).
var explicitMime = map[string]string{
	".mp3":  "audio/mpeg",
	".flac": "audio/flac",
	".wav":  "audio/wav",
	".m4a":  "audio/mp4",
	".ogg":  "audio/ogg",
}

// startMediaServer binds an HTTP server to a random loopback port and serves
// the media handler in the background. It returns the base URL (e.g.
// "http://127.0.0.1:52341") the frontend should prepend to track URLs.
func startMediaServer(lib *Library) (string, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}
	srv := &http.Server{Handler: newMediaServer(lib)}
	go func() { _ = srv.Serve(ln) }()

	addr := ln.Addr().(*net.TCPAddr)
	return fmt.Sprintf("http://127.0.0.1:%d", addr.Port), nil
}

func (s *mediaServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// The media server runs on a different origin than the WebView, so audio,
	// cover and lyric requests are cross-origin. Permit them and expose the
	// range headers the media element relies on for seeking.
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Expose-Headers", "Accept-Ranges, Content-Range, Content-Length")
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	path := r.URL.Path
	switch {
	case strings.HasPrefix(path, "/media/"):
		s.serveMedia(w, r, strings.TrimPrefix(path, "/media/"))
	case strings.HasPrefix(path, "/cover/"):
		s.serveCover(w, r, strings.TrimPrefix(path, "/cover/"))
	case strings.HasPrefix(path, "/lyric/"):
		s.serveLyric(w, r, strings.TrimPrefix(path, "/lyric/"))
	default:
		http.NotFound(w, r)
	}
}

// serveMedia streams the audio file. http.ServeContent transparently handles
// HTTP range requests, which the WebView needs for seeking within a track.
func (s *mediaServer) serveMedia(w http.ResponseWriter, r *http.Request, id string) {
	t, ok := s.lib.get(id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	f, err := os.Open(t.Path)
	if err != nil {
		http.Error(w, "cannot open media", http.StatusNotFound)
		return
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		http.Error(w, "cannot stat media", http.StatusInternalServerError)
		return
	}

	if mime, ok := explicitMime[strings.ToLower(filepath.Ext(t.Path))]; ok {
		w.Header().Set("Content-Type", mime)
	}
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Cache-Control", "no-cache")
	http.ServeContent(w, r, info.Name(), info.ModTime(), f)
}

func (s *mediaServer) serveCover(w http.ResponseWriter, r *http.Request, id string) {
	t, ok := s.lib.get(id)
	if !ok || len(t.coverData) == 0 {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", t.coverMime)
	w.Header().Set("Cache-Control", "public, max-age=86400")
	http.ServeContent(w, r, "cover", time.Time{}, bytes.NewReader(t.coverData))
}

func (s *mediaServer) serveLyric(w http.ResponseWriter, r *http.Request, id string) {
	t, ok := s.lib.get(id)
	if !ok || t.LyricURL == "" || t.lyricFile == "" {
		http.NotFound(w, r)
		return
	}
	data, err := os.ReadFile(t.lyricFile)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	_, _ = w.Write(data)
}
