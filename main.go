package main

import (
	"bytes"
	"embed"
	_ "embed"
	"encoding/json"
	"html/template"
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

//go:embed assets/* templates/*
var webAssets embed.FS

// TemplateData holds common data passed to all templates
type TemplateData struct {
	Title            string
	Page             string
	HeroTitle        string
	HeroDescription  string
	ContentID        string
	ScriptFile       string
}

var templates *template.Template

func init() {
	// Load all templates from embedded filesystem
	var err error
	templates, err = template.ParseFS(webAssets, "templates/base.html", "templates/tts.html", "templates/stt.html")
	if err != nil {
		panic("Failed to load templates: " + err.Error())
	}
}

func main() {
	// Create a new Gin router with default middleware
	router := gin.Default()

	// Serve static files from embedded filesystem at /assets/
	// Use fs.Sub to serve from assets/ subdirectory
	assetsFS, _ := fs.Sub(webAssets, "assets")
	router.StaticFS("/assets", http.FS(assetsFS))

	// Serve the home page
	router.GET("/", serveHome)

	// Serve the speech-to-text page
	router.GET("/stt", serveSTT)

	// TTS endpoint that calls speaches.ai server
	router.POST("/api/tts", handleTTS)

	// STT endpoint for speech-to-text requests
	router.POST("/api/stt", handleSTT)

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

// serveHome renders the Text-to-Speech page using templates
func serveHome(c *gin.Context) {
	data := TemplateData{
		Title:            "🍑 Speaches UI",
		Page:             "tts",
		HeroTitle:        "👄 Text-to-Speech",
		HeroDescription:  "Convert text to natural-sounding speech with multiple voices and models",
		ContentID:        "tts",
	}

	c.Header("Content-Type", "text/html; charset=utf-8")

	// Render base.html with tts.html content template included
	if err := templates.ExecuteTemplate(c.Writer, "base.html", data); err != nil {
		// ERROR: Failed to render TTS template
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to render page"})
		return
	}
}

// serveSTT renders the Speech-to-Text page using templates
func serveSTT(c *gin.Context) {
	data := TemplateData{
		Title:            "🍑 Speaches UI - Speech to Text",
		Page:             "stt",
		HeroTitle:        "👂 Speech-to-Text",
		HeroDescription:  "Convert speech to text with advanced transcription models",
		ContentID:        "stt",
	}

	c.Header("Content-Type", "text/html; charset=utf-8")

	// Render base.html with stt.html content template included
	if err := templates.ExecuteTemplate(c.Writer, "base.html", data); err != nil {
		// ERROR: Failed to render STT template
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to render page"})
		return
	}
}

// handleSTT processes speech-to-text requests by calling the speaches.ai server
func handleSTT(c *gin.Context) {
	// Get language and model from form data
	language := c.DefaultPostForm("language", "en")
	model := c.DefaultPostForm("model", "standard")

	// Get the audio file from the form
	file, err := c.FormFile("audio")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "audio file is required"})
		return
	}

	// Validate language
	validLanguages := map[string]bool{
		"en": true, "es": true, "fr": true, "de": true, "it": true,
		"pt": true, "ja": true, "ko": true, "zh": true,
	}
	if !validLanguages[language] {
		language = "en"
	}

	// Validate model
	validModels := map[string]bool{
		"fast": true, "standard": true, "accurate": true,
	}
	if !validModels[model] {
		model = "standard"
	}

	// Read the audio file
	src, err := file.Open()
	if err != nil {
		// ERROR: Failed to open uploaded audio file
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to open audio file"})
		return
	}
	defer src.Close()

	audioData, err := io.ReadAll(src)
	if err != nil {
		// ERROR: Failed to read audio file data
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read audio file"})
		return
	}

	// Create multipart request for speaches.ai
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add audio file to multipart request (field name must be "file")
	part, err := writer.CreateFormFile("file", file.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create form file"})
		return
	}

	_, err = part.Write(audioData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write audio data"})
		return
	}

	// Add language field
	writer.WriteField("language", language)

	// Add model field - map quality to a model identifier
	modelValue := "whisper-1" // default model
	writer.WriteField("model", modelValue)

	writer.Close()

	// Call the speaches.ai server
	speachesBaseURL := os.Getenv("SPEACHES_URL")
	if speachesBaseURL == "" {
		speachesBaseURL = "http://localhost:8000"
	}
	speachesURL := speachesBaseURL + "/v1/audio/transcriptions"

	req, err := http.NewRequest("POST", speachesURL, body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
		return
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		// ERROR: Failed to connect to speaches.ai server
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "speaches.ai server is not available. Make sure it's running on localhost:8000"})
		return
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		errorMsg := string(bodyBytes)

		// Check if error is about missing model and try to download it
		if bytes.Contains(bodyBytes, []byte("is not installed locally")) || (bytes.Contains(bodyBytes, []byte("Model")) && bytes.Contains(bodyBytes, []byte("not found"))) {
			// Try to download the model
			downloadURL := speachesBaseURL + "/v1/models/whisper-1"
			downloadResp, downloadErr := http.Post(downloadURL, "application/json", nil)
			if downloadErr == nil {
				downloadResp.Body.Close()

				// Retry the transcription request after downloading
				// Recreate the request body since the previous one was consumed
				body2 := &bytes.Buffer{}
				writer2 := multipart.NewWriter(body2)

				part2, _ := writer2.CreateFormFile("file", file.Filename)
				part2.Write(audioData)

				writer2.WriteField("language", language)
				writer2.WriteField("model", "whisper-1")
				writer2.Close()

				req2, err2 := http.NewRequest("POST", speachesURL, body2)
				if err2 == nil {
					req2.Header.Set("Content-Type", writer2.FormDataContentType())

					resp2, err3 := client.Do(req2)
					if err3 == nil {
						defer resp2.Body.Close()

						if resp2.StatusCode == http.StatusOK {
							// Success! Parse and return the response
							var result struct {
								Text string `json:"text"`
							}

							json.NewDecoder(resp2.Body).Decode(&result)
							c.JSON(http.StatusOK, gin.H{"text": result.Text})
							return
						}
					}
				}
			}
		}

		// If we get here, return the original error
		// ERROR: speaches.ai server returned an error
		c.JSON(resp.StatusCode, gin.H{"error": "speaches.ai server error: " + errorMsg})
		return
	}

	// Parse the response
	var result struct {
		Text string `json:"text"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		// ERROR: Failed to decode speaches.ai response
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode transcription response"})
		return
	}

	// Return the transcribed text
	c.JSON(http.StatusOK, gin.H{"text": result.Text})
}
