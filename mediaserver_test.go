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
