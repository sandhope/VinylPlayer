# Vinyl Player

**English** | [中文](./README.zh-CN.md)

A desktop audio player built with [Wails](https://wails.io/), powered by a Go backend and a Vue 3 frontend. It features a vinyl-record-inspired interface, a built-in equalizer, spectrum visualization, and synchronized lyrics — just point it at your local music and press play.

## ✨ Features

- 🎵 **Multi-format support**: MP3, FLAC, WAV, M4A, OGG (decoded natively by WebView2)
- 🏷️ **Automatic metadata**: reads embedded title, artist, album, and cover art
- 🎚️ **6-band equalizer**: built-in presets (Flat, Pop, Rock, Jazz, Classical, Bass, Vocal) plus custom gains
- 📊 **Real-time spectrum visualization**: dynamic audio spectrum powered by the Web Audio API
- 📜 **Synchronized lyrics**: auto-loads `.lrc` files matching the audio file name and highlights lines during playback
- 🎨 **Multiple themes**: Retro, Dark, Light, and Minimal
- 🔀 **Playback controls**: shuffle, repeat one/all, volume, and seek
- 📁 **Local media library**: import folders or individual files; the library is persisted and restored on restart
- 🖼️ **Frameless window**: custom title bar and window controls

## 📸 Screenshots

<table>
  <tr>
    <td align="center">
      <img src="screenshots/dark_retro.png" width="400"/><br/>
      <sub>Retro (Dark)</sub>
    </td>
    <td align="center">
      <img src="screenshots/dark_immersed.png" width="400"/><br/>
      <sub>Immersive (Dark)</sub>
    </td>
  </tr>
  <tr>
    <td align="center">
      <img src="screenshots/light_modern.png" width="400"/><br/>
      <sub>Modern (Light)</sub>
    </td>
    <td align="center">
      <img src="screenshots/light_simple.png" width="400"/><br/>
      <sub>Simple (Light)</sub>
    </td>
  </tr>
</table>

## 🏗️ Architecture

| Layer | Technology |
| --- | --- |
| Desktop framework | Wails v2 |
| Backend | Go 1.25 |
| Frontend | Vue 3 + Vite |
| Metadata parsing | [dhowden/tag](https://github.com/dhowden/tag) |
| Audio processing | Web Audio API (equalizer + spectrum) |

Backend modules:

- `main.go` — application entry point; configures the Wails window and starts the media server
- `app.go` — app logic bound to the frontend (library loading, file pickers, window controls)
- `library.go` — music library management: scans files, extracts metadata, builds Tracks
- `mediaserver.go` — a dedicated loopback HTTP server that streams audio, cover art, and lyrics via Range requests
- `store.go` — library persistence (saved to `%AppData%/VinylPlayer/library.json`)

> Note: media assets run on a separate loopback port (rather than the Wails asset server) so that streaming and seeking work correctly in both `wails dev` and production builds.

## 🚀 Getting Started

### Prerequisites

- [Go](https://go.dev/) 1.25+
- [Node.js](https://nodejs.org/) (with npm)
- [Wails CLI](https://wails.io/docs/gettingstarted/installation): `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- On Windows, the WebView2 runtime is required (usually pre-installed on Win10/11)

### Development

Run in the project root for live frontend reload:

```bash
wails dev
```

To debug in a browser and call Go methods, open http://localhost:34115.

### Build

Produce a redistributable, production build:

```bash
wails build
```

The output is `VinylPlayer.exe`.

## 📖 Usage

1. On first launch, the bundled `audios/` sample tracks are loaded automatically.
2. Import a local music folder or one/multiple audio files from the UI; the library is saved automatically.
3. Place a `.lrc` lyric file in the same directory with the same name as the audio file to enable synchronized lyrics.
4. Choose an equalizer preset or manually adjust the 6 bands in the equalizer panel.
5. Switch the interface style anytime via the theme switcher.

## 📂 Project Structure

```
VinylPlayer/
├── app.go              # App logic bound to the frontend
├── main.go             # Application entry point
├── library.go          # Music library & metadata parsing
├── mediaserver.go      # Local media streaming server
├── store.go            # Library persistence
├── audios/             # Sample audio
└── frontend/           # Vue 3 frontend
    └── src/
        ├── components/     # UI components (player, lyrics, equalizer, etc.)
        └── composables/    # Playback logic & theming (usePlayer / useTheme)
```

## 📄 Configuration

You can configure the project by editing `wails.json`. See the [Wails project configuration docs](https://wails.io/docs/reference/project-config) for details.
