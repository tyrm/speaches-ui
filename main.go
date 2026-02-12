package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// Create a new Gin router with default middleware
	router := gin.Default()

	// Serve the home page
	router.GET("/", serveHome)

	// Start the server on port 5420
	// INFO: Server listening on http://localhost:5420
	router.Run(":5420")
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
		}
		.btn-speak:hover {
			background: linear-gradient(135deg, #764ba2 0%, #667eea 100%);
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
		</form>
	</div>

	<!-- Bootstrap 5.3 JS from CDN -->
	<script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/js/bootstrap.bundle.min.js"></script>
	<script>
		// Handle the speak button click
		document.getElementById('speakBtn').addEventListener('click', function() {
			// TODO: Implement text-to-speech functionality
			console.log('Speak button clicked');
			console.log('Text:', document.getElementById('paragraphInput').value);
		});
	</script>
</body>
</html>
	`
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}
