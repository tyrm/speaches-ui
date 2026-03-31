# Speaches UI

[![Claude Logo](https://img.shields.io/badge/Claude-D97757?label=generated%20with)](https://claude.ai/code)

A Text-To-Speech web application built with Go (Gin framework) and Bootstrap 5.3. Integrates with the speaches.ai server for TTS processing.

## Quick Start

### Requirements
- Go 1.21+
- speaches.ai server running on localhost:8000 (or custom `SPEACHES_URL`)

### Setup & Run

```bash
# Install dependencies
go mod tidy

# Start speaches.ai server (separate terminal)
# See speaches.ai documentation

# Run this application
go run main.go
```

Visit `http://localhost:5420`

## Configuration

Set the `SPEACHES_URL` environment variable to use a custom speaches.ai server:
```bash
set SPEACHES_URL=http://example.com:8000  # Windows
export SPEACHES_URL=http://example.com:8000  # Linux/macOS
```

Default: `http://localhost:8000`

## Usage

### Text-to-Speech

1. Enter text in the textarea
2. Select a **Model** (Kokoro or Piper)
3. Select a **Voice** (varies by model)
4. Choose an **Output Format**: MP3, WAV, FLAC, or PCM
5. Adjust **Speed**: 0.25× to 4.0× (1.0× is normal)
6. Set **Sample Rate**: 8000–48000 Hz (default 24000 Hz — higher = better quality, larger file)
7. Click **Speak** or press **Shift+Enter**
8. Audio plays automatically in the player
9. Click **⬇ Download** to save the audio file

Your selections are automatically saved and restored on the next visit!

## Supported Voices

### Kokoro (Neural TTS) - Model `tts-1`

**🇺🇸 American English:**
- **Female (11):** Nova, Sarah, Bella, Heart, Aoede, Jessica, Kore, Nicole, River, Sky, Alloy
- **Male (9):** Adam, Echo, Liam, Onyx, Michael, Eric, Fenrir, Puck, Santa

**🇬🇧 British English:**
- **Female (4):** Alice, Emma, Isabella, Lily
- **Male (4):** Fable, George, Daniel, Lewis

**🇪🇸 Spanish:**
- **Female (1):** Dora
- **Male (2):** Alex, Santa

**🇫🇷 French:**
- **Female (1):** Siwis

**🇮🇳 Hindi:**
- **Female (2):** Alpha, Beta
- **Male (2):** Omega, Psi

**🇮🇹 Italian:**
- **Female (1):** Sara
- **Male (1):** Nicola

**🇧🇷 Brazilian Portuguese:**
- **Female (1):** Dora
- **Male (2):** Alex, Santa

**🇯🇵 Japanese:**
- **Female (4):** Alpha, Gongitsune, Nezumi, Tebukuro
- **Male (1):** Kumo

**🇨🇳 Mandarin Chinese:**
- **Female (4):** Xiaobei, Xiaoni, Xiaoxiao, Xiaoyi
- **Male (4):** Yunjian, Yunxi, Yunxia, Yunyang

### Piper (Fast TTS) - Model `tts-1-piper`

**🇺🇸 US English:**
- Ryan (3 quality levels: high, medium, low)
- Female voices: HFC Female, Amy, Kathleen, Kristin, LJ Speech
- Male voices: HFC Male, Lessac, Danny, Joe, John, Bryce, Kusal, Norman, and others
- Specialized: LibriTTS, Arctic, L2Arctic variants

**🇬🇧 British English:**
- Alan, Southern English Female, Alba, Aru, Cori, Jenny Dioco, Northern English Male, Semaine, VCTK

## API

### POST `/api/tts`

Generate speech from text.

**Request:**
```json
{
  "text": "Your text here",
  "model": "tts-1",
  "voice": "af_nova",
  "format": "mp3",
  "speed": 1.0,
  "sample_rate": 24000
}
```

**Parameters:**
- `text` (string, required): Text to convert to speech
- `model` (string, optional): `tts-1` (Kokoro) or `tts-1-piper` (Piper). Default: `tts-1`
- `voice` (string, optional): Voice ID (varies by model)
- `format` (string, optional): Output format — `mp3`, `wav`, `flac`, or `pcm`. Default: `mp3`
- `speed` (float, optional): Speech rate from 0.25× to 4.0×. Default: `1.0`
- `sample_rate` (int, optional): Audio sample rate in Hz, range 8000–48000. Default: `24000`

**Response:** Audio stream in the specified format, or error JSON

**Example:**
```bash
curl -X POST http://localhost:5420/api/tts \
  -H "Content-Type: application/json" \
  -d '{"text":"Hello world","format":"wav","speed":1.5,"sample_rate":48000}' \
  --output speech.wav
```

## Project Structure

```
├── main.go                      # Server, routes, and API handlers
├── assets/
│   ├── css/
│   │   ├── bootstrap.min.css    # Bootstrap 5.3 framework
│   │   └── style.css            # Shared application styles
│   ├── js/
│   │   └── bootstrap.bundle.min.js
│   ├── index.html               # Legacy (kept for reference)
│   └── stt.html                 # Legacy (kept for reference)
├── templates/
│   ├── base.html                # Base template with shared layout
│   ├── tts.html                 # Text-to-Speech page content
│   └── stt.html                 # Speech-to-Text page content
├── go.mod                        # Go dependencies
└── README.md                     # This file
```

## Template System

The application uses Go's `html/template` package with a unified template architecture:

- **base.html**: Shared HTML structure (doctype, head, navbar, hero section, footer)
- **style.css**: Centralized styles for consistent formatting across all pages
- **tts.html** & **stt.html**: Page-specific content templates that extend base.html

This approach ensures:
✅ Consistent UI/UX across all pages
✅ Single source of truth for styles
✅ Easy to maintain navbar and layout changes
✅ DRY principle - no duplicated HTML

## License

MIT
