# Awesome TTS Project

[![Claude Logo](https://img.shields.io/badge/Claude-D97757?label=generated%20with)](https://claude.ai/code)

A Text-To-Speech (TTS) web application built with Go and the Gin framework.

## Features

- 🌐 Web-based interface using Bootstrap 5.3
- 🚀 Fast and lightweight Gin web server
- 📝 Text input for paragraph entry
- 🔊 Speak button (functionality to be implemented)

## Requirements

- Go 1.21 or later
- Modern web browser

## Installation & Setup

1. Clone the repository or navigate to the project directory
2. Install dependencies:
   ```bash
   go mod tidy
   ```

## Running the Server

Start the server:
```bash
go run main.go
```

The server will start on `http://localhost:5420`

## Project Structure

- `main.go` - Main server application with Gin router and HTML handler
- `go.mod` - Go module dependencies
- `README.md` - This file

## Current Status

The foundation is set up with:
- ✅ Web server listening on port 5420
- ✅ HTML interface with Bootstrap styling
- ✅ Text input and speak button UI
- ⏳ TTS functionality (coming soon)

## Development Notes

This is a foundation project with the main loop intentionally kept simple for expansion. The speak button is currently a placeholder and will implement text-to-speech functionality in future updates.

## License

MIT
