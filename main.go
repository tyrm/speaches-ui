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
		Text string `json:"text" binding:"required"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "text field is required"})
		return
	}

	if req.Text == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "text cannot be empty"})
		return
	}

	// Create request payload for speaches.ai server (OpenAI API compatible)
	payload := map[string]interface{}{
		"model": "tts-1",
		"input": req.Text,
		"voice": "alloy",
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
			background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
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
			background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
			border: none;
			cursor: pointer;
			transition: all 0.3s ease;
		}
		.btn-speak:hover {
			background: linear-gradient(135deg, #764ba2 0%, #667eea 100%);
			transform: translateY(-2px);
			box-shadow: 0 5px 15px rgba(102, 126, 234, 0.4);
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
			<button type="button" class="btn btn-primary btn-speak" id="speakBtn">
				Speak
			</button>
			<div id="statusMessage"></div>
			<div id="errorAlert" class="alert alert-danger" role="alert"></div>
			<div id="successAlert" class="alert alert-success" role="alert"></div>
		</form>

		<!-- Hidden audio element for playback -->
		<audio id="audioPlayer" style="display: none;"></audio>
	</div>

	<!-- Bootstrap 5.3 JS from CDN -->
	<script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/js/bootstrap.bundle.min.js"></script>
	<script>
		const speakBtn = document.getElementById('speakBtn');
		const textInput = document.getElementById('paragraphInput');
		const audioPlayer = document.getElementById('audioPlayer');
		const errorAlert = document.getElementById('errorAlert');
		const successAlert = document.getElementById('successAlert');
		const statusMessage = document.getElementById('statusMessage');

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
				// Send request to TTS endpoint
				const response = await fetch('/api/tts', {
					method: 'POST',
					headers: {
						'Content-Type': 'application/json',
					},
					body: JSON.stringify({ text: text })
				});

				if (!response.ok) {
					const errorData = await response.json();
					throw new Error(errorData.error || 'Failed to generate speech');
				}

				// Get the audio blob from the response
				const audioBlob = await response.blob();
				const audioUrl = URL.createObjectURL(audioBlob);

				// Play the audio
				audioPlayer.src = audioUrl;
				audioPlayer.play();

				showSuccess('Speech generated and playing');
				statusMessage.textContent = '';

				// Clean up the URL when audio ends
				audioPlayer.addEventListener('ended', function() {
					URL.revokeObjectURL(audioUrl);
					statusMessage.textContent = '';
				}, { once: true });

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
