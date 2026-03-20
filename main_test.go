package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// --- newTranslateFunc tests ---

func TestNewTranslateFunc_KnownKey_English(t *testing.T) {
	localizer := i18n.NewLocalizer(i18nBundle, "en")
	translate := newTranslateFunc(localizer)

	got := translate("nav_tts")
	assert.Equal(t, "Text to Speech", got)
}

func TestNewTranslateFunc_KnownKey_Spanish(t *testing.T) {
	localizer := i18n.NewLocalizer(i18nBundle, "es")
	translate := newTranslateFunc(localizer)

	got := translate("nav_tts")
	assert.Equal(t, "Texto a Voz", got)
}

func TestNewTranslateFunc_UnknownKey_ReturnsKeyAsIs(t *testing.T) {
	localizer := i18n.NewLocalizer(i18nBundle, "en")
	translate := newTranslateFunc(localizer)

	// Unknown keys must degrade gracefully — return the message ID itself.
	got := translate("this_key_does_not_exist")
	assert.Equal(t, "this_key_does_not_exist", got)
}

func TestNewTranslateFunc_FallsBackToEnglish_ForUnsupportedLocale(t *testing.T) {
	// Japanese is not in our supported locales; should fall back to English.
	localizer := i18n.NewLocalizer(i18nBundle, "ja")
	translate := newTranslateFunc(localizer)

	got := translate("nav_stt")
	assert.Equal(t, "Speech to Text", got)
}

func TestNewTranslateFunc_AllRequiredKeysExist(t *testing.T) {
	localizer := i18n.NewLocalizer(i18nBundle, "en")
	translate := newTranslateFunc(localizer)

	requiredKeys := []string{
		// Navigation
		"nav_tts", "nav_stt", "nav_models",
		"toggle_dark_mode", "theme_light", "theme_dark", "theme_system",
		"lang_switcher_label",
		// TTS
		"tts_enter_text", "tts_placeholder", "tts_select_model", "tts_loading_models",
		"tts_select_voice", "tts_speak_button", "tts_no_voices", "tts_error_loading_voices",
		"tts_error_loading_models", "tts_no_tts_models", "tts_enter_text_error",
		"tts_speaking", "tts_generating", "tts_failed", "tts_no_audio", "tts_play", "tts_pause",
		// STT
		"stt_choose_file", "stt_no_file", "stt_transcribed_text", "stt_placeholder",
		"stt_select_language", "lang_english", "lang_spanish", "lang_french",
		"lang_german", "lang_italian", "lang_portuguese", "lang_japanese",
		"lang_korean", "lang_chinese", "stt_select_model", "stt_fast", "stt_standard",
		"stt_accurate", "stt_transcribe_button", "stt_transcribing", "stt_processing",
		"stt_success", "stt_audio_loaded", "stt_play", "stt_pause",
		// Models
		"models_refresh", "models_add_tts", "models_add_stt", "models_loading",
		"models_tts_section", "models_stt_section", "models_no_tts", "models_no_stt",
		"models_installed_badge", "models_id_label", "models_desc_label", "models_type_label",
		// Add models
		"back_to_models", "search_tts_placeholder", "search_stt_placeholder",
		"table_model_name", "table_model_id", "table_description", "table_status", "table_action",
		"status_installed", "status_not_installed", "install_button", "installing_button",
		"registry_loading", "registry_no_tts", "registry_no_stt",
	}

	for _, key := range requiredKeys {
		t.Run(key, func(t *testing.T) {
			result := translate(key)
			// The translation must not fall back to the key itself for English.
			assert.NotEmpty(t, result, "translation for %q should not be empty", key)
			assert.NotEqual(t, key, result, "key %q has no English translation (returned key itself)", key)
		})
	}
}

// --- i18nMiddleware tests ---

func newTestRouter() *gin.Engine {
	r := gin.New()
	r.Use(i18nMiddleware())
	r.GET("/lang", func(c *gin.Context) {
		lang := langFromCtx(c)
		c.String(http.StatusOK, lang)
	})
	return r
}

func TestI18nMiddleware_DefaultsToEnglish(t *testing.T) {
	r := newTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/lang", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "en", rec.Body.String())
}

func TestI18nMiddleware_LangQueryParam(t *testing.T) {
	r := newTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/lang?lang=es", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "es", rec.Body.String())
}

func TestI18nMiddleware_AcceptLanguageHeader(t *testing.T) {
	r := newTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/lang", nil)
	req.Header.Set("Accept-Language", "es-ES,es;q=0.9")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "es", rec.Body.String())
}

func TestI18nMiddleware_QueryParamOverridesHeader(t *testing.T) {
	r := newTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/lang?lang=en", nil)
	req.Header.Set("Accept-Language", "es-ES,es;q=0.9")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	// ?lang=en should win over the Accept-Language header
	assert.Equal(t, "en", rec.Body.String())
}

func TestI18nMiddleware_UnsupportedLangFallsToEnglish(t *testing.T) {
	r := newTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/lang?lang=zh", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	// zh is not in our supported locales; must fall back to English
	assert.Equal(t, "en", rec.Body.String())
}

// --- localizerFromCtx / langFromCtx tests ---

func TestLocalizerFromCtx_NoLocalizerSet_ReturnsEnglishLocalizer(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	// Don't set "localizer" in the context — simulates a missing middleware.
	localizer := localizerFromCtx(c)
	require.NotNil(t, localizer)

	translate := newTranslateFunc(localizer)
	got := translate("nav_tts")
	assert.Equal(t, "Text to Speech", got)
}

func TestLangFromCtx_NoLangSet_ReturnsEn(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	lang := langFromCtx(c)
	assert.Equal(t, "en", lang)
}

// --- Template rendering smoke tests ---

func newFullRouter() *gin.Engine {
	r := gin.New()
	r.Use(i18nMiddleware())
	r.GET("/", serveHome)
	r.GET("/stt", serveSTT)
	r.GET("/models", serveModels)
	r.GET("/add-tts-models", serveAddTTSModels)
	r.GET("/add-stt-models", serveAddSTTModels)
	return r
}

func TestServeHome_RendersOK(t *testing.T) {
	r := newFullRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.True(t, strings.Contains(body, "Text to Speech"), "expected nav link 'Text to Speech' in body")
}

func TestServeHome_SpanishTranslation(t *testing.T) {
	r := newFullRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/?lang=es", nil)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.True(t, strings.Contains(body, "Texto a Voz"), "expected Spanish nav link in body")
}

func TestServeSTT_RendersOK(t *testing.T) {
	r := newFullRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/stt", nil)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.True(t, strings.Contains(body, "Transcription will appear here"), "expected STT placeholder text in body")
}

func TestServeSTT_SpanishTranslation(t *testing.T) {
	r := newFullRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/stt?lang=es", nil)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.True(t, strings.Contains(body, "Transcribir"), "expected Spanish transcribe button in body")
}

func TestServeModels_RendersOK(t *testing.T) {
	r := newFullRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/models", nil)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.True(t, strings.Contains(body, "Text-to-Speech Models"), "expected TTS section heading in body")
}

func TestServeAddTTSModels_RendersOK(t *testing.T) {
	r := newFullRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/add-tts-models", nil)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.True(t, strings.Contains(body, "Back to Models"), "expected back link in body")
}

func TestServeAddSTTModels_RendersOK(t *testing.T) {
	r := newFullRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/add-stt-models", nil)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.True(t, strings.Contains(body, "Search STT models"), "expected STT search placeholder in body")
}

// --- formatModelName tests ---

func TestFormatModelName(t *testing.T) {
	tests := []struct {
		name    string
		modelID string
		want    string
	}{
		{
			name:    "kokoro model",
			modelID: "tts-1",
			want:    "Kokoro (Neural TTS)",
		},
		{
			name:    "piper ryan high",
			modelID: "speaches-ai/piper-en_US-ryan-high",
			want:    "Piper - Ryan (TTS)",
		},
		{
			name:    "whisper model",
			modelID: "whisper-1",
			want:    "Whisper v1 (Speech to Text)",
		},
		{
			name:    "unknown model formatted",
			modelID: "my-custom-model",
			want:    "My Custom Model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatModelName(tt.modelID)
			assert.Equal(t, tt.want, got)
		})
	}
}

// --- isSTTModel tests ---

func TestIsSTTModel(t *testing.T) {
	tests := []struct {
		name    string
		modelID string
		want    bool
	}{
		{name: "whisper model", modelID: "whisper-1", want: true},
		{name: "speech model", modelID: "speech-model-v2", want: true},
		{name: "transcription model", modelID: "transcription-en", want: true},
		{name: "kokoro tts model", modelID: "tts-1", want: false},
		{name: "piper model", modelID: "speaches-ai/piper-en_US-ryan-high", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSTTModel(tt.modelID)
			assert.Equal(t, tt.want, got)
		})
	}
}

// --- i18nBundle integrity: Spanish locale covers all English keys ---

func TestSpanishLocaleCoversAllEnglishKeys(t *testing.T) {
	enLocalizer := i18n.NewLocalizer(i18nBundle, "en")
	esLocalizer := i18n.NewLocalizer(i18nBundle, "es")

	enT := newTranslateFunc(enLocalizer)
	esT := newTranslateFunc(esLocalizer)

	// Spot-check a representative set of keys that are explicitly translated
	// in es.toml. If a key degrades to the message ID, we know it's missing.
	translatedInES := []string{
		"nav_tts", "nav_stt", "nav_models",
		"tts_speak_button", "tts_play", "tts_pause",
		"stt_transcribe_button", "stt_transcribing", "stt_success",
		"models_refresh", "models_add_tts", "models_add_stt",
		"back_to_models", "install_button", "installing_button",
		"status_installed", "status_not_installed",
	}

	for _, key := range translatedInES {
		t.Run(key, func(t *testing.T) {
			enVal := enT(key)
			esVal := esT(key)
			// The Spanish value must differ from the English one (it's translated).
			assert.NotEqual(t, enVal, esVal,
				"key %q: Spanish value %q matches English %q — translation may be missing",
				key, esVal, enVal)
			// And it must not fall back to the key itself.
			assert.NotEqual(t, key, esVal,
				"key %q has no Spanish translation (returned key itself)", key)
		})
	}
}

// --- language.MatchStrings helper test ---

func TestLanguageMatcher_MatchesSpanish(t *testing.T) {
	matcher := language.NewMatcher([]language.Tag{
		language.English,
		language.Spanish,
	})

	tag, _ := language.MatchStrings(matcher, "es-ES,es;q=0.9")
	base, _ := tag.Base()
	assert.Equal(t, "es", base.String())
}

func TestLanguageMatcher_FallsBackToEnglish(t *testing.T) {
	matcher := language.NewMatcher([]language.Tag{
		language.English,
		language.Spanish,
	})

	tag, _ := language.MatchStrings(matcher, "zh-CN")
	base, _ := tag.Base()
	assert.Equal(t, "en", base.String())
}
