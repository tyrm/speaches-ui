package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// Create a new Gin router with default middleware
	router := gin.Default()

	// Serve the home page
	router.GET("/", serveHome)

	// TTS endpoint that calls speaches.ai server
	router.POST("/api/tts", handleTTS)

	// Start the server on port 5420
	// INFO: Server listening on http://localhost:5420
	router.Run(":5420")
}

// handleTTS processes text-to-speech requests by calling the speaches.ai server
func handleTTS(c *gin.Context) {
	var req struct {
		Text  string `json:"text" binding:"required"`
		Voice string `json:"voice"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "text field is required"})
		return
	}

	if req.Text == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "text cannot be empty"})
		return
	}

	// Set default voice if not provided
	voice := req.Voice
	if voice == "" {
		voice = "alloy"
	}

	// Validate voice is one of the available voices (Kokoro model)
	validVoices := map[string]bool{
		// American Female
		"af_nova":  true,
		"af_sarah": true,
		"af_bella": true,
		"af_heart": true,
		// American Male
		"am_adam":  true,
		"am_echo":  true,
		"am_liam":  true,
		"am_onyx":  true,
		// British
		"bf_alice":  true,
		"bm_fable":  true,
		"bm_george": true,
	}
	if !validVoices[voice] {
		// WARNING: Invalid voice requested, using default
		voice = "af_nova"
	}

	// Create request payload for speaches.ai server (OpenAI API compatible)
	payload := map[string]interface{}{
		"model": "tts-1",
		"input": req.Text,
		"voice": voice,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal request"})
		return
	}

	// Call the speaches.ai server on localhost:8000
	speachesURL := "http://localhost:8000/v1/audio/speech"
	resp, err := http.Post(speachesURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		// ERROR: Failed to connect to speaches.ai server on localhost:8000
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "speaches.ai server is not available. Make sure it's running on localhost:8000"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// ERROR: speaches.ai server returned an error
		body, _ := io.ReadAll(resp.Body)
		c.JSON(resp.StatusCode, gin.H{"error": "speaches.ai server error: " + string(body)})
		return
	}

	// Set proper audio response headers
	c.Header("Content-Type", "audio/mpeg")
	c.Header("Content-Disposition", "inline")

	// Stream the audio response back to the client
	io.Copy(c.Writer, resp.Body)
}

// serveHome serves the main HTML page
func serveHome(c *gin.Context) {
	html := `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Awesome TTS Project</title>
	<!-- Bootstrap 5.3 CSS from CDN -->
	<link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
	<style>
		body {
			display: flex;
			align-items: center;
			justify-content: center;
			min-height: 100vh;
			background: linear-gradient(135deg, #0052b3 0%, #004d73 100%);
		}
		.container-main {
			background: white;
			padding: 40px;
			border-radius: 10px;
			box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
			max-width: 600px;
			width: 100%;
		}
		h1 {
			color: #333;
			margin-bottom: 30px;
			text-align: center;
		}
		.form-group {
			margin-bottom: 20px;
		}
		label {
			font-weight: 600;
			color: #555;
			margin-bottom: 10px;
		}
		.form-control {
			border-color: #ddd;
			transition: border-color 0.3s ease, box-shadow 0.3s ease;
		}
		.form-control:focus {
			border-color: #0052b3;
			box-shadow: 0 0 0 0.2rem rgba(0, 82, 179, 0.25);
		}
		select.form-control {
			cursor: pointer;
			padding: 8px 12px;
			font-size: 14px;
		}
		select.form-control option {
			padding: 8px 12px;
		}
		textarea.form-control {
			resize: vertical;
			min-height: 400px !important;
			height: 400px !important;
			font-size: 16px;
		}
		.btn-speak {
			width: 100%;
			padding: 12px;
			font-size: 18px;
			font-weight: 600;
			background: linear-gradient(135deg, #0052b3 0%, #004d73 100%);
			border: none;
			cursor: pointer;
			transition: all 0.3s ease;
		}
		.btn-speak:hover {
			background: linear-gradient(135deg, #004d73 0%, #0052b3 100%);
			transform: translateY(-2px);
			box-shadow: 0 5px 15px rgba(0, 82, 179, 0.4);
		}
		.btn-speak:disabled {
			opacity: 0.6;
			cursor: not-allowed;
			transform: none;
		}
		.alert {
			margin-top: 20px;
			display: none;
		}
		.alert.show {
			display: block;
		}
		#statusMessage {
			text-align: center;
			margin-top: 15px;
			font-size: 14px;
			color: #666;
			min-height: 20px;
		}
		.player-container {
			display: none;
			background: #f8f9fa;
			border: 1px solid #dee2e6;
			border-radius: 8px;
			padding: 15px;
			margin-top: 20px;
		}
		.player-container.show {
			display: block;
		}
		.player-controls {
			display: flex;
			align-items: center;
			gap: 10px;
			margin-bottom: 10px;
		}
		.play-btn {
			background: linear-gradient(135deg, #0052b3 0%, #004d73 100%);
			color: white;
			border: none;
			padding: 8px 16px;
			border-radius: 5px;
			cursor: pointer;
			font-size: 14px;
			font-weight: 600;
			transition: all 0.3s ease;
			min-width: 80px;
		}
		.play-btn:hover {
			transform: translateY(-2px);
			box-shadow: 0 4px 10px rgba(0, 82, 179, 0.3);
		}
		.play-btn:active {
			transform: translateY(0);
		}
		.progress-container {
			flex: 1;
			display: flex;
			align-items: center;
			gap: 10px;
		}
		.progress-bar {
			flex: 1;
			height: 6px;
			background: #dee2e6;
			border-radius: 3px;
			cursor: pointer;
			appearance: none;
			-webkit-appearance: none;
			width: 100%;
		}
		.progress-bar::-webkit-slider-thumb {
			appearance: none;
			-webkit-appearance: none;
			width: 14px;
			height: 14px;
			background: linear-gradient(135deg, #0052b3 0%, #004d73 100%);
			border-radius: 50%;
			cursor: pointer;
			box-shadow: 0 2px 5px rgba(0, 0, 0, 0.2);
		}
		.progress-bar::-moz-range-thumb {
			width: 14px;
			height: 14px;
			background: linear-gradient(135deg, #0052b3 0%, #004d73 100%);
			border-radius: 50%;
			cursor: pointer;
			border: none;
			box-shadow: 0 2px 5px rgba(0, 0, 0, 0.2);
		}
		.time-display {
			font-size: 12px;
			color: #666;
			white-space: nowrap;
			min-width: 50px;
			text-align: right;
		}
	</style>
</head>
<body>
	<div class="container-main">
		<h1>🔊 Awesome TTS Project</h1>
		<form id="ttsForm">
			<div class="form-group">
				<label for="paragraphInput">Enter text to speak:</label>
				<textarea
					class="form-control"
					id="paragraphInput"
					placeholder="Type your text here..."
					required></textarea>
			</div>
			<div class="form-group">
				<label for="voiceSelect">Select Voice:</label>
				<select class="form-control" id="voiceSelect">
					<optgroup label="American Female">
						<option value="af_nova">Nova (Neutral)</option>
						<option value="af_sarah">Sarah (Clear)</option>
						<option value="af_bella">Bella (Warm)</option>
						<option value="af_heart">Heart (Expressive)</option>
					</optgroup>
					<optgroup label="American Male">
						<option value="am_adam">Adam (Friendly)</option>
						<option value="am_echo">Echo (Deep)</option>
						<option value="am_liam">Liam (Professional)</option>
						<option value="am_onyx">Onyx (Commanding)</option>
					</optgroup>
					<optgroup label="British">
						<option value="bf_alice">Alice (Female)</option>
						<option value="bm_fable">Fable (Male)</option>
						<option value="bm_george">George (Male)</option>
					</optgroup>
				</select>
			</div>
			<button type="button" class="btn btn-primary btn-speak" id="speakBtn">
				Speak
			</button>
			<div id="statusMessage"></div>
			<div id="errorAlert" class="alert alert-danger" role="alert"></div>
			<div id="successAlert" class="alert alert-success" role="alert"></div>
		</form>

		<!-- Audio Player -->
		<div id="playerContainer" class="player-container">
			<div class="player-controls">
				<button id="playBtn" class="play-btn">▶ Play</button>
				<div class="progress-container">
					<input type="range" id="progressBar" class="progress-bar" min="0" value="0">
					<span id="timeDisplay" class="time-display">0:00 / 0:00</span>
				</div>
			</div>
		</div>

		<!-- Hidden audio element for playback -->
		<audio id="audioPlayer"></audio>
	</div>

	<!-- Bootstrap 5.3 JS from CDN -->
	<script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/js/bootstrap.bundle.min.js"></script>
	<script>
		const speakBtn = document.getElementById('speakBtn');
		const textInput = document.getElementById('paragraphInput');
		const voiceSelect = document.getElementById('voiceSelect');
		const audioPlayer = document.getElementById('audioPlayer');
		const playerContainer = document.getElementById('playerContainer');
		const playBtn = document.getElementById('playBtn');
		const progressBar = document.getElementById('progressBar');
		const timeDisplay = document.getElementById('timeDisplay');
		const errorAlert = document.getElementById('errorAlert');
		const successAlert = document.getElementById('successAlert');
		const statusMessage = document.getElementById('statusMessage');

		let audioUrl = null;

		// Handle the speak button click
		speakBtn.addEventListener('click', async function() {
			const text = textInput.value.trim();

			// Validate input
			if (!text) {
				showError('Please enter some text to speak');
				return;
			}

			// Disable button and show loading state
			speakBtn.disabled = true;
			speakBtn.textContent = '🔊 Speaking...';
			statusMessage.textContent = 'Generating speech...';
			hideAllAlerts();
			playerContainer.classList.remove('show');

			try {
				// Send request to TTS endpoint with selected voice
				const response = await fetch('/api/tts', {
					method: 'POST',
					headers: {
						'Content-Type': 'application/json',
					},
					body: JSON.stringify({
						text: text,
						voice: voiceSelect.value
					})
				});

				if (!response.ok) {
					const errorData = await response.json();
					throw new Error(errorData.error || 'Failed to generate speech');
				}

				// Get the audio blob from the response
				const audioBlob = await response.blob();
				if (audioUrl) {
					URL.revokeObjectURL(audioUrl);
				}
				audioUrl = URL.createObjectURL(audioBlob);

				// Set audio source and show player
				audioPlayer.src = audioUrl;
				playerContainer.classList.add('show');
				resetPlayer();
				audioPlayer.play();
				updatePlayButton();

				showSuccess('Speech generated and playing');
				statusMessage.textContent = '';

			} catch (error) {
				console.error('TTS Error:', error);
				showError('Error: ' + error.message);
				statusMessage.textContent = '';
				playerContainer.classList.remove('show');
			} finally {
				// Re-enable button
				speakBtn.disabled = false;
				speakBtn.textContent = '🔊 Speak';
			}
		});

		// Play/Pause button handler
		playBtn.addEventListener('click', function() {
			if (audioPlayer.paused) {
				audioPlayer.play();
			} else {
				audioPlayer.pause();
			}
			updatePlayButton();
		});

		// Update play button text
		audioPlayer.addEventListener('play', updatePlayButton);
		audioPlayer.addEventListener('pause', updatePlayButton);

		function updatePlayButton() {
			if (audioPlayer.paused) {
				playBtn.textContent = '▶ Play';
			} else {
				playBtn.textContent = '⏸ Pause';
			}
		}

		// Progress bar seeking
		progressBar.addEventListener('input', function() {
			if (audioPlayer.duration) {
				audioPlayer.currentTime = (progressBar.value / 100) * audioPlayer.duration;
			}
		});

		// Update progress bar and time display as audio plays
		audioPlayer.addEventListener('timeupdate', function() {
			if (audioPlayer.duration) {
				progressBar.value = (audioPlayer.currentTime / audioPlayer.duration) * 100;
			}
			updateTimeDisplay();
		});

		// Update progress bar max when metadata loads
		audioPlayer.addEventListener('loadedmetadata', function() {
			progressBar.max = 100;
			updateTimeDisplay();
		});

		// Reset player when audio ends
		audioPlayer.addEventListener('ended', function() {
			resetPlayer();
		});

		function resetPlayer() {
			progressBar.value = 0;
			updatePlayButton();
			updateTimeDisplay();
		}

		function updateTimeDisplay() {
			const current = formatTime(audioPlayer.currentTime);
			const duration = formatTime(audioPlayer.duration);
			timeDisplay.textContent = current + ' / ' + duration;
		}

		function formatTime(seconds) {
			if (!seconds || isNaN(seconds)) return '0:00';
			const mins = Math.floor(seconds / 60);
			const secs = Math.floor(seconds % 60);
			return mins + ':' + secs.toString().padStart(2, '0');
		}

		function showError(message) {
			errorAlert.textContent = message;
			errorAlert.classList.add('show');
			successAlert.classList.remove('show');
		}

		function showSuccess(message) {
			successAlert.textContent = message;
			successAlert.classList.add('show');
			errorAlert.classList.remove('show');
		}

		function hideAllAlerts() {
			errorAlert.classList.remove('show');
			successAlert.classList.remove('show');
		}

		// Allow Enter key to submit (Shift+Enter)
		textInput.addEventListener('keydown', function(e) {
			if (e.shiftKey && e.key === 'Enter') {
				e.preventDefault();
				speakBtn.click();
			}
		});
	</script>
</body>
</html>
	`
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}
