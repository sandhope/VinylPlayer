package main

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestMediaServerServesAudio(t *testing.T) {
	lib := NewLibrary()
	tracks := lib.scanDirectory("audios")
	if len(tracks) == 0 {
		t.Skip("no sample audios found")
	}
	base, err := startMediaServer(lib)
	if err != nil {
		t.Fatalf("startMediaServer: %v", err)
	}

	tr := tracks[0]
	resp, err := http.Get(base + tr.URL)
	if err != nil {
		t.Fatalf("GET media: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "audio/") {
		t.Fatalf("Content-Type = %q, want audio/*", ct)
	}
	if resp.Header.Get("Accept-Ranges") != "bytes" {
		t.Fatalf("Accept-Ranges missing")
	}
	if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Fatalf("CORS header missing")
	}
	n, _ := io.Copy(io.Discard, resp.Body)
	if n == 0 {
		t.Fatalf("empty body")
	}
	t.Logf("served %q -> %s, %d bytes", tr.Title, ct, n)

	// Range request (seeking) should yield 206 Partial Content.
	req, _ := http.NewRequest(http.MethodGet, base+tr.URL, nil)
	req.Header.Set("Range", "bytes=0-1023")
	rr, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("range GET: %v", err)
	}
	defer rr.Body.Close()
	if rr.StatusCode != http.StatusPartialContent {
		t.Fatalf("range status = %d, want 206", rr.StatusCode)
	}
	t.Logf("range request OK: %s", rr.Header.Get("Content-Range"))
}

// TestMediaServerServesLyrics verifies that sibling .lrc files are discovered
// even when they carry an artist suffix (e.g. "水中花 - 谭咏麟.lrc" next to
// "水中花.mp3") and that the media server returns their contents.
func TestMediaServerServesLyrics(t *testing.T) {
	lib := NewLibrary()
	tracks := lib.scanDirectory("audios")
	if len(tracks) == 0 {
		t.Skip("no sample audios found")
	}
	base, err := startMediaServer(lib)
	if err != nil {
		t.Fatalf("startMediaServer: %v", err)
	}

	var withLyrics *Track
	for _, tr := range tracks {
		if tr.LyricURL != "" {
			withLyrics = tr
			break
		}
	}
	if withLyrics == nil {
		t.Fatal("no track resolved a sibling .lrc file")
	}

	resp, err := http.Get(base + withLyrics.LyricURL)
	if err != nil {
		t.Fatalf("GET lyric: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if len(body) == 0 {
		t.Fatalf("empty lyric body")
	}
	if !strings.Contains(string(body), "[") {
		t.Fatalf("lyric body does not look like LRC: %.40q", string(body))
	}
	t.Logf("served lyrics for %q -> %d bytes", withLyrics.Title, len(body))
}
