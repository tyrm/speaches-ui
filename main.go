package main

import (
	"bytes"
	"embed"
	_ "embed"
	"encoding/json"
	"io"
	"io/fs"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

//go:embed assets/*
var webAssets embed.FS

func main() {
	// Create a new Gin router with default middleware
	router := gin.Default()

	// Serve static files from embedded filesystem at /assets/
	// Use fs.Sub to serve from assets/ subdirectory
	assetsFS, _ := fs.Sub(webAssets, "assets")
	router.StaticFS("/assets", http.FS(assetsFS))

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
		Model string `json:"model"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "text field is required"})
		return
	}

	if req.Text == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "text cannot be empty"})
		return
	}

	// Set default model if not provided
	model := req.Model
	if model == "" {
		model = "tts-1"
	}

	// Set default voice if not provided
	voice := req.Voice

	// Validate voice based on model
	kokoroVoices := map[string]bool{
		// American Female
		"af_nova":   true,
		"af_sarah":  true,
		"af_bella":  true,
		"af_heart":  true,
		"af_aoede":  true,
		"af_jessica": true,
		"af_kore":   true,
		"af_nicole": true,
		"af_river":  true,
		"af_sky":    true,
		"af_alloy":  true,
		// American Male
		"am_adam":    true,
		"am_echo":    true,
		"am_liam":    true,
		"am_onyx":    true,
		"am_michael": true,
		"am_eric":    true,
		"am_fenrir":  true,
		"am_puck":    true,
		"am_santa":   true,
		// British Female
		"bf_alice":     true,
		"bf_emma":      true,
		"bf_isabella":  true,
		"bf_lily":      true,
		// British Male
		"bm_fable":  true,
		"bm_george": true,
		"bm_daniel": true,
		"bm_lewis":  true,
	}

	piperVoices := map[string]bool{
		// English US - Ryan
		"en_US-ryan-high":   true,
		"en_US-ryan-low":    true,
		"en_US-ryan-medium": true,
		// English US - Female
		"en_US-amy-low":           true,
		"en_US-amy-medium":        true,
		"en_US-hfc_female-medium": true,
		"en_US-kathleen-low":      true,
		"en_US-kristin-medium":    true,
		"en_US-ljspeech-high":     true,
		"en_US-ljspeech-medium":   true,
		// English US - Male
		"en_US-hfc_male-medium": true,
		"en_US-lessac-high":     true,
		"en_US-lessac-low":      true,
		"en_US-lessac-medium":   true,
		"en_US-danny-low":       true,
		"en_US-joe-medium":      true,
		"en_US-john-medium":     true,
		"en_US-bryce-medium":    true,
		"en_US-kusal-medium":    true,
		"en_US-norman-medium":   true,
		// English US - Other
		"en_US-libritts-high":     true,
		"en_US-libritts_r-medium": true,
		"en_US-arctic-medium":     true,
		"en_US-l2arctic-medium":   true,
		// English GB
		"en_GB-alan-low":                     true,
		"en_GB-alan-medium":                  true,
		"en_GB-southern_english_female-low":  true,
		"en_GB-alba-medium":                  true,
		"en_GB-aru-medium":                   true,
		"en_GB-cori-high":                    true,
		"en_GB-cori-medium":                  true,
		"en_GB-jenny_dioco-medium":           true,
		"en_GB-northern_english_male-medium": true,
		"en_GB-semaine-medium":               true,
		"en_GB-vctk-medium":                  true,
	}

	// Validate and set defaults based on model
	var actualModel string
	if model == "tts-1" {
		if !kokoroVoices[voice] {
			voice = "af_nova"
		}
		actualModel = "tts-1"
	} else if model == "tts-1-piper" {
		if !piperVoices[voice] {
			voice = "en_US-ryan-medium"
		}
		// For Piper, the model is the full path: speaches-ai/piper-{voice}
		actualModel = "speaches-ai/piper-" + voice
	} else {
		// Unknown model, default to Kokoro
		model = "tts-1"
		voice = "af_nova"
		actualModel = "tts-1"
	}

	// Create request payload for speaches.ai server (OpenAI API compatible)
	payload := map[string]interface{}{
		"model": actualModel,
		"input": req.Text,
		"voice": voice,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal request"})
		return
	}

	// Call the speaches.ai server using SPEACHES_URL environment variable
	speachesBaseURL := os.Getenv("SPEACHES_URL")
	if speachesBaseURL == "" {
		speachesBaseURL = "http://localhost:8000"
	}
	speachesURL := speachesBaseURL + "/v1/audio/speech"

	// Try to make the TTS request
	resp, err := http.Post(speachesURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		// ERROR: Failed to connect to speaches.ai server on localhost:8000
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "speaches.ai server is not available. Make sure it's running on localhost:8000"})
		return
	}
	defer resp.Body.Close()

	// Check if model needs to be downloaded
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		errorMsg := string(body)

		// Check if error is about missing model (for Piper voices)
		if model == "tts-1-piper" && (bytes.Contains(body, []byte("is not installed locally")) || (bytes.Contains(body, []byte("Model")) && bytes.Contains(body, []byte("not found")))) {
			// Auto-download the Piper voice model
			// URL-encode the model ID for the download endpoint
			modelID := "speaches-ai%2Fpiper-" + voice
			downloadURL := speachesBaseURL + "/v1/models/" + modelID
			downloadResp, downloadErr := http.Post(downloadURL, "application/json", nil)
			if downloadErr == nil {
				downloadResp.Body.Close()

				// Retry the TTS request after downloading
				resp2, err2 := http.Post(speachesURL, "application/json", bytes.NewBuffer(jsonPayload))
				if err2 != nil {
					c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Failed to generate speech after downloading model"})
					return
				}
				defer resp2.Body.Close()

				if resp2.StatusCode == http.StatusOK {
					// Success! Stream the audio
					c.Header("Content-Type", "audio/mpeg")
					c.Header("Content-Disposition", "inline")
					io.Copy(c.Writer, resp2.Body)
					return
				}
			}
		}

		// If we get here, return the original error
		c.JSON(resp.StatusCode, gin.H{"error": "speaches.ai server error: " + errorMsg})
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
	<title>🍑 Speaches UI</title>
	<!-- Bootstrap 5.3 CSS from local assets -->
	<link href="/assets/css/bootstrap.min.css" rel="stylesheet">
	<style>
		body {
			display: flex;
			align-items: center;
			justify-content: center;
			min-height: 100vh;
			background: linear-gradient(135deg, #0052b3 0%, #00a870 100%);
			padding: 60px 20px;
		}
		.container-main {
			background: white;
			padding: 40px;
			border-radius: 10px;
			box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
			max-width: 1200px;
			width: 100%;
		}
		.form-layout {
			display: flex;
			gap: 30px;
			align-items: flex-start;
		}
		.form-left {
			flex: 1;
		}
		.form-right {
			width: 300px;
			flex-shrink: 0;
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
			border-color: #00a870;
			box-shadow: 0 0 0 0.2rem rgba(0, 168, 112, 0.25);
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
			background: linear-gradient(135deg, #0052b3 0%, #00a870 100%);
			border: none;
			cursor: pointer;
			transition: all 0.3s ease;
		}
		.btn-speak:hover {
			background: linear-gradient(135deg, #00a870 0%, #0052b3 100%);
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
			display: block;
			background: #f8f9fa;
			border: 1px solid #dee2e6;
			border-radius: 8px;
			padding: 15px;
			margin-bottom: 20px;
		}
		.player-controls {
			display: flex;
			align-items: center;
			gap: 10px;
			margin-bottom: 10px;
		}
		.play-btn {
			background: linear-gradient(135deg, #0052b3 0%, #00a870 100%);
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
			background: linear-gradient(135deg, #0052b3 0%, #00a870 100%);
			border-radius: 50%;
			cursor: pointer;
			box-shadow: 0 2px 5px rgba(0, 0, 0, 0.2);
		}
		.progress-bar::-moz-range-thumb {
			width: 14px;
			height: 14px;
			background: linear-gradient(135deg, #0052b3 0%, #00a870 100%);
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
		<h1>🍑 Speaches UI</h1>

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

		<form id="ttsForm">
			<div class="form-layout">
				<div class="form-left">
					<div class="form-group">
						<label for="paragraphInput">Enter text to speak:</label>
						<textarea
							class="form-control"
							id="paragraphInput"
							placeholder="Type your text here..."
							required></textarea>
					</div>
				</div>
				<div class="form-right">
					<div class="form-group">
						<label for="modelSelect">Select Model:</label>
						<select class="form-control" id="modelSelect">
							<option value="tts-1">Kokoro (Neural TTS)</option>
							<option value="tts-1-piper">Piper (Fast TTS)</option>
						</select>
					</div>
					<div class="form-group">
						<label for="voiceSelect">Select Voice:</label>
						<select class="form-control" id="voiceSelect">
							<!-- Voices will be populated dynamically based on model -->
						</select>
					</div>
					<button type="button" class="btn btn-primary btn-speak" id="speakBtn">
						Speak
					</button>
					<div id="statusMessage"></div>
				<div id="errorAlert" class="alert alert-danger" role="alert"></div>
				<div id="successAlert" class="alert alert-success" role="alert"></div>
				</div>
			</div>
		</form>
	</div>

	<!-- Bootstrap 5.3 JS from local assets -->
	<script src="/assets/js/bootstrap.bundle.min.js"></script>
	<script>
		const speakBtn = document.getElementById('speakBtn');
		const textInput = document.getElementById('paragraphInput');
		const modelSelect = document.getElementById('modelSelect');
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

		// Voice options for each model
		const voiceOptions = {
			'tts-1': {
				'American Female': [
					{ value: 'af_nova', label: 'Nova (Neutral)' },
					{ value: 'af_sarah', label: 'Sarah (Clear)' },
					{ value: 'af_bella', label: 'Bella (Warm)' },
					{ value: 'af_heart', label: 'Heart (Expressive)' },
					{ value: 'af_aoede', label: 'Aoede (Bright)' },
					{ value: 'af_jessica', label: 'Jessica (Smooth)' },
					{ value: 'af_kore', label: 'Kore (Dynamic)' },
					{ value: 'af_nicole', label: 'Nicole (Natural)' },
					{ value: 'af_river', label: 'River (Calm)' },
					{ value: 'af_sky', label: 'Sky (Gentle)' },
					{ value: 'af_alloy', label: 'Alloy (Balanced)' }
				],
				'American Male': [
					{ value: 'am_adam', label: 'Adam (Friendly)' },
					{ value: 'am_echo', label: 'Echo (Deep)' },
					{ value: 'am_liam', label: 'Liam (Professional)' },
					{ value: 'am_onyx', label: 'Onyx (Commanding)' },
					{ value: 'am_michael', label: 'Michael (Energetic)' },
					{ value: 'am_eric', label: 'Eric (Smooth)' },
					{ value: 'am_fenrir', label: 'Fenrir (Intense)' },
					{ value: 'am_puck', label: 'Puck (Playful)' },
					{ value: 'am_santa', label: 'Santa (Jolly)' }
				],
				'British Female': [
					{ value: 'bf_alice', label: 'Alice (Posh)' },
					{ value: 'bf_emma', label: 'Emma (Refined)' },
					{ value: 'bf_isabella', label: 'Isabella (Elegant)' },
					{ value: 'bf_lily', label: 'Lily (Sweet)' }
				],
				'British Male': [
					{ value: 'bm_fable', label: 'Fable (Theatrical)' },
					{ value: 'bm_george', label: 'George (Distinguished)' },
					{ value: 'bm_daniel', label: 'Daniel (Smooth)' },
					{ value: 'bm_lewis', label: 'Lewis (Rich)' }
				]
			},
			'tts-1-piper': {
				'Ryan (US Male)': [
					{ value: 'en_US-ryan-high', label: 'High Quality' },
					{ value: 'en_US-ryan-medium', label: 'Medium Quality' },
					{ value: 'en_US-ryan-low', label: 'Low Quality' }
				],
				'US Female Voices': [
					{ value: 'en_US-hfc_female-medium', label: 'HFC Female' },
					{ value: 'en_US-amy-medium', label: 'Amy Medium' },
					{ value: 'en_US-amy-low', label: 'Amy Low' },
					{ value: 'en_US-kathleen-low', label: 'Kathleen' },
					{ value: 'en_US-kristin-medium', label: 'Kristin' },
					{ value: 'en_US-ljspeech-high', label: 'LJ Speech High' },
					{ value: 'en_US-ljspeech-medium', label: 'LJ Speech Medium' }
				],
				'US Male Voices': [
					{ value: 'en_US-hfc_male-medium', label: 'HFC Male' },
					{ value: 'en_US-lessac-high', label: 'Lessac High' },
					{ value: 'en_US-lessac-medium', label: 'Lessac Medium' },
					{ value: 'en_US-lessac-low', label: 'Lessac Low' },
					{ value: 'en_US-danny-low', label: 'Danny' },
					{ value: 'en_US-joe-medium', label: 'Joe' },
					{ value: 'en_US-john-medium', label: 'John' },
					{ value: 'en_US-bryce-medium', label: 'Bryce' },
					{ value: 'en_US-norman-medium', label: 'Norman' }
				],
				'British Voices': [
					{ value: 'en_GB-alan-medium', label: 'Alan Medium' },
					{ value: 'en_GB-alan-low', label: 'Alan Low' },
					{ value: 'en_GB-southern_english_female-low', label: 'Southern Female' },
					{ value: 'en_GB-alba-medium', label: 'Alba' },
					{ value: 'en_GB-cori-high', label: 'Cori High' },
					{ value: 'en_GB-cori-medium', label: 'Cori Medium' },
					{ value: 'en_GB-jenny_dioco-medium', label: 'Jenny Dioco' }
				]
			}
		};

		// Populate voice dropdown based on selected model
		function updateVoiceOptions() {
			const selectedModel = modelSelect.value;
			const voices = voiceOptions[selectedModel];

			// Remember currently selected voice before clearing
			const previousVoice = voiceSelect.value;

			// Clear current options
			voiceSelect.innerHTML = '';

			// Add new options based on model
			for (const [group, voiceList] of Object.entries(voices)) {
				const optgroup = document.createElement('optgroup');
				optgroup.label = group;

				voiceList.forEach(voice => {
					const option = document.createElement('option');
					option.value = voice.value;
					option.textContent = voice.label;
					optgroup.appendChild(option);
				});

				voiceSelect.appendChild(optgroup);
			}

			// Try to restore previous voice if it exists in new model
			if (previousVoice && Array.from(voiceSelect.options).some(opt => opt.value === previousVoice)) {
				voiceSelect.value = previousVoice;
			}
		}

		// Load saved preferences from localStorage
		function loadPreferences() {
			const savedModel = localStorage.getItem('tts-model');
			const savedVoice = localStorage.getItem('tts-voice');

			if (savedModel) {
				modelSelect.value = savedModel;
			}

			return savedVoice;
		}

		// Save model preference
		function saveModelPreference() {
			localStorage.setItem('tts-model', modelSelect.value);
		}

		// Save voice preference
		function saveVoicePreference() {
			localStorage.setItem('tts-voice', voiceSelect.value);
		}

		// Load preferences and initialize
		const savedVoice = loadPreferences();
		updateVoiceOptions();

		// Restore saved voice after options are populated
		if (savedVoice && Array.from(voiceSelect.options).some(opt => opt.value === savedVoice)) {
			voiceSelect.value = savedVoice;
		}

		// Update voice options when model changes
		modelSelect.addEventListener('change', function() {
			saveModelPreference();
			updateVoiceOptions();
		});

		// Save voice preference when it changes
		voiceSelect.addEventListener('change', saveVoicePreference);

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

			try {
				// Send request to TTS endpoint with selected model and voice
				const response = await fetch('/api/tts', {
					method: 'POST',
					headers: {
						'Content-Type': 'application/json',
					},
					body: JSON.stringify({
						text: text,
						model: modelSelect.value,
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

				// Set audio source and reset player
				audioPlayer.src = audioUrl;
				resetPlayer();
				audioPlayer.play();
				updatePlayButton();

					statusMessage.textContent = '';

			} catch (error) {
				console.error('TTS Error:', error);
				showError('Error: ' + error.message);
				statusMessage.textContent = '';
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
