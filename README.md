# Speaches UI

[![Claude Logo](https://img.shields.io/badge/Claude-D97757?label=generated%20with)](https://claude.ai/code)

A Text-To-Speech and Speech-to-Text web application built with Go (Gin framework) and Bootstrap 5.3. Integrates with the speaches.ai server for TTS/STT processing. Supports multiple languages via go-i18n.

## Quick Start

### Requirements
- Go 1.24+
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
export SPEACHES_URL=http://example.com:8000  # Linux/macOS
set SPEACHES_URL=http://example.com:8000     # Windows
```

Default: `http://localhost:8000`

## Internationalization (i18n)

The UI supports multiple languages powered by [go-i18n](https://github.com/nicksnyder/go-i18n). Locale files are TOML and live in `locales/`.

### Switching Language

**Query parameter (highest priority):**
```
http://localhost:5420/?lang=es
http://localhost:5420/stt?lang=es
```

**Accept-Language HTTP header:**
The server respects the browser's `Accept-Language` header automatically.

**In-page language switcher:**
A dropdown in the navbar lets users switch between available languages. The selection propagates via query parameter on navigation links.

### Supported Languages

| Code | Language  |
|------|-----------|
| `en` | English (default) |
| `es` | Español   |

### Adding a New Language

1. Create `locales/<code>.toml` using `locales/en.toml` as the reference.
2. Translate all message keys. Untranslated keys fall back to English automatically.
3. Register the new language tag in the `matcher` slice inside `i18nMiddleware()` in `main.go`.
4. Add an entry for the new language to the navbar dropdown in `templates/base.html`.

## Usage

1. Enter text in the textarea
2. Select a model:
   - **Kokoro**: High-quality neural TTS with English voices (American & British)
   - **Piper**: Fast TTS with English and Spanish voices
3. Select a voice from the dropdown
4. Click **Speak** or press **Shift+Enter**
5. Audio plays automatically

### Supported Languages & Voices

**Kokoro (TTS-1)**
- American: Female (Nova, Sarah, Bella, Heart, Aoede, Jessica, Kore, Nicole, River, Sky, Alloy) | Male (Adam, Echo, Liam, Onyx, Michael, Eric, Fenrir, Puck, Santa, Lewis)
- British: Female (Alice, Emma, Isabella, Lily) | Male (Fable, George, Daniel, Lewis)

**Piper (TTS-1-Piper)**
- US English: Multiple male and female voice variants at different quality levels
- British English: Alan, Alba, Aru, Cori, Jenny Dioco, and more
- Spanish: Carlfm, Davefx, Sharvard, MLS 10246, MLS 9972

## API

### POST `/api/tts`

Generate speech from text.

**Request:**
```json
{
  "text": "Your text here",
  "voice": "af_nova"
}
```

**Response:** Audio stream (audio/mpeg) or error JSON

**Example:**
```bash
curl -X POST http://localhost:5420/api/tts \
  -H "Content-Type: application/json" \
  -d '{"text":"Hello"}' --output speech.mp3
```

### POST `/api/stt`

Transcribe audio to text (multipart form).

**Form fields:** `audio` (file), `language` (e.g. `en`), `model` (`fast`|`standard`|`accurate`)

**Response:** `{"text": "transcribed text"}`

### GET `/api/models`

List installed TTS and STT models.

### GET `/api/models/registry`

List all models available in the registry with install status.

### POST `/api/models/install`

Install a model by ID: `{"model_id": "tts-1"}`

### GET `/api/voices/:modelId`

List available voices for a given model.

## Project Structure

```
├── main.go                          # Server, routes, API handlers, i18n setup
├── main_test.go                     # Unit and integration tests
├── locales/
│   ├── en.toml                      # English strings (source of truth)
│   └── es.toml                      # Spanish translations
├── assets/
│   ├── css/
│   │   ├── bootstrap.min.css        # Bootstrap 5.3 framework
│   │   └── style.css                # Shared application styles
│   └── js/
│       └── bootstrap.bundle.min.js
├── templates/
│   ├── base.html                    # Base layout: navbar, hero, language switcher
│   ├── tts.html                     # Text-to-Speech page content
│   ├── stt.html                     # Speech-to-Text page content
│   ├── models.html                  # Installed models page
│   ├── add-tts-models.html          # Add TTS models from registry
│   └── add-stt-models.html          # Add STT models from registry
├── go.mod
└── README.md
```

## Template System

The application uses Go's `html/template` package with a unified template architecture:

- **base.html**: Shared HTML structure (doctype, head, navbar with language switcher, hero section)
- **style.css**: Centralized styles for consistent formatting
- **Page templates**: Content templates that extend base.html via `{{template "...-content" .}}`

Each template receives a `TemplateData` struct with a `T` function field for translations:

```html
<!-- Static HTML strings -->
<label>{{call .T "tts_enter_text"}}</label>

<!-- JavaScript strings via inline i18n object -->
<script>
const i18n = {
  tts_play: "{{call .T "tts_play"}}",
  tts_pause: "{{call .T "tts_pause"}}",
};
</script>
```

## Development

### Running Tests

```bash
# Run all tests with race detector
go test -race ./...

# Run with verbose output
go test -race -v ./...

# Run with coverage
go test -cover ./...
```

### Adding Translations

All message IDs are defined in `locales/en.toml`. Each entry uses the go-i18n TOML format:

```toml
[my_new_key]
description = "Human-readable description for translators"
other = "English text"
```

Use `{{call .T "my_new_key"}}` in templates or populate the `i18n` JS object for use in scripts.

## License

MIT
