package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed translations/*.json
var translationsFS embed.FS

// Translation holds all translation strings
type Translation map[string]interface{}

var translations = make(map[string]Translation)

// SupportedLanguages lists all available languages
var SupportedLanguages = []string{"en", "de"}

// DefaultLanguage is the fallback language
const DefaultLanguage = "en"

// Init loads all translation files
func Init() error {
	for _, lang := range SupportedLanguages {
		data, err := translationsFS.ReadFile(fmt.Sprintf("translations/%s.json", lang))
		if err != nil {
			return fmt.Errorf("failed to load translation for %s: %w", lang, err)
		}

		var t Translation
		if err := json.Unmarshal(data, &t); err != nil {
			return fmt.Errorf("failed to parse translation for %s: %w", lang, err)
		}

		translations[lang] = t
	}
	return nil
}

// GetTranslation returns the translation map for a given language
func GetTranslation(lang string) Translation {
	// Normalize language code (e.g., "en-US" -> "en")
	lang = strings.ToLower(strings.Split(lang, "-")[0])

	if t, ok := translations[lang]; ok {
		return t
	}
	// Fallback to default language
	return translations[DefaultLanguage]
}

// T is a helper function to get a translation value by key path
func T(lang, keyPath string) string {
	t := GetTranslation(lang)
	keys := strings.Split(keyPath, ".")

	var current interface{} = map[string]interface{}(t)
	for _, key := range keys {
		if m, ok := current.(map[string]interface{}); ok {
			if val, exists := m[key]; exists {
				current = val
			} else {
				// Key not found, return the key path as fallback
				return keyPath
			}
		} else {
			// Not a map, return the key path as fallback
			return keyPath
		}
	}

	if str, ok := current.(string); ok {
		return str
	}

	return keyPath
}

// GetLanguageFromAcceptHeader parses the Accept-Language header
func GetLanguageFromAcceptHeader(acceptLang string) string {
	if acceptLang == "" {
		return DefaultLanguage
	}

	// Parse the Accept-Language header (simplified)
	// Format: "en-US,en;q=0.9,de;q=0.8"
	langs := strings.Split(acceptLang, ",")
	for _, lang := range langs {
		// Remove quality factor if present
		lang = strings.TrimSpace(strings.Split(lang, ";")[0])
		// Normalize to just the language code
		langCode := strings.ToLower(strings.Split(lang, "-")[0])

		// Check if we support this language
		for _, supported := range SupportedLanguages {
			if langCode == supported {
				return supported
			}
		}
	}

	return DefaultLanguage
}
