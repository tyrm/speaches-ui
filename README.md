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

1. Enter text in the textarea
2. Select a voice from the dropdown
3. Click **Speak** or press **Shift+Enter**
4. Audio plays automatically

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
