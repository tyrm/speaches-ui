# Awesome TTS Project

[![Claude Logo](https://img.shields.io/badge/Claude-D97757?label=generated%20with)](https://claude.ai/code)

A Text-To-Speech (TTS) web application built with Go and the Gin framework.

## Features

- 🌐 Web-based interface using Bootstrap 5.3
- 🚀 Fast and lightweight Gin web server
- 📝 Large text input box (400px) for paragraph entry
- 🔊 Text-to-Speech with speaches.ai server integration
- ⚡ Auto-playing audio with smooth error handling
- 🎤 OpenAI API compatible TTS endpoint

## Requirements

- Go 1.21 or later
- Modern web browser with Web Audio API support
- **speaches.ai server** running on `localhost:8000` (OpenAI API compatible TTS server)

## Installation & Setup

1. Clone the repository or navigate to the project directory
2. Install dependencies:
   ```bash
   go mod tidy
   ```

## Running the Server

Start the speaches.ai server on port 8000 (make sure it's running first):
```bash
# In another terminal, start speaches.ai server
# Refer to speaches.ai documentation for installation
```

Then start this application:
```bash
go run main.go
```

The server will start on `http://localhost:5420`

## Usage

1. Open your browser to `http://localhost:5420`
2. Enter text in the text box (minimum 1 character)
3. Click the **Speak** button or press **Shift+Enter**
4. The text will be sent to speaches.ai server for TTS processing
5. Audio will be generated and automatically played in your browser

**Features:**
- Real-time error messages if speaches.ai server is unavailable
- Loading state while generating speech
- Keyboard shortcut: Shift+Enter to speak
- Smooth animations and responsive design

## Project Structure

- `main.go` - Main server application with Gin router and HTML handler
- `go.mod` - Go module dependencies
- `README.md` - This file

## Current Status

Full TTS implementation complete:
- ✅ Web server listening on port 5420
- ✅ Large 400px textarea for text input
- ✅ Bootstrap 5.3 responsive UI with animations
- ✅ Backend TTS endpoint (`POST /api/tts`)
- ✅ speaches.ai server integration (OpenAI API compatible)
- ✅ Auto-playing audio in browser
- ✅ Error handling and user feedback
- ✅ Keyboard shortcut support (Shift+Enter)

## API Documentation

### POST `/api/tts`
Generates speech from text using the speaches.ai server.

**Request:**
```json
{
  "text": "Your text here"
}
```

**Response:**
- **200 OK**: Audio stream (audio/mpeg)
- **400 Bad Request**: Missing or empty text field
- **503 Service Unavailable**: speaches.ai server not available

**Example:**
```bash
curl -X POST http://localhost:5420/api/tts \
  -H "Content-Type: application/json" \
  -d '{"text":"Hello world"}' \
  --output speech.mp3
```

## Development Notes

This project integrates with speaches.ai for TTS functionality. The backend uses the OpenAI API-compatible endpoint (`/v1/audio/speech`) running on localhost:8000.

**Configuration:**
- **Server Port:** 5420
- **speaches.ai Port:** 8000 (configurable in `main.go` line ~50)
- **TTS Model:** tts-1
- **Voice:** alloy

## License

MIT
