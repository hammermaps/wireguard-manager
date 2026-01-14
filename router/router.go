package router

import (
	"errors"
	"io"
	"io/fs"
	"net/http"
	"strings"
	"text/template"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/swissmakers/wireguard-manager/handler"
	"github.com/swissmakers/wireguard-manager/i18n"
	"github.com/swissmakers/wireguard-manager/util"
)

// TemplateRegistry is a custom html/template renderer for the Echo framework.
type TemplateRegistry struct {
	templates map[string]*template.Template
	extraData map[string]interface{}
}

// Render implements the e.Renderer interface.
// It injects extra data into the template if data is a map.
func (t *TemplateRegistry) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	tmpl, ok := t.templates[name]
	if !ok {
		return errors.New("Template not found -> " + name)
	}

	// Inject extra app data if data is a map.
	if m, ok := data.(map[string]interface{}); ok {
		for k, v := range t.extraData {
			m[k] = v
		}
		m["client_defaults"] = util.ClientDefaultsFromEnv()

		// Detect language from cookie or Accept-Language header
		lang := i18n.DefaultLanguage
		if cookie, err := c.Cookie("language"); err == nil && cookie.Value != "" {
			lang = cookie.Value
		} else {
			acceptLang := c.Request().Header.Get("Accept-Language")
			lang = i18n.GetLanguageFromAcceptHeader(acceptLang)
		}

		// Add translations and current language to template data
		m["lang"] = lang
		translation := i18n.GetTranslation(lang)
		m["t"] = translation

		log.Debugf("Rendering template %s with language %s (translation keys: %d)", name, lang, len(translation))
	}

	// For the login page, no base layout is needed.
	if name == "login.html" {
		return tmpl.Execute(w, data)
	}

	return tmpl.Execute(w, data)
}

// New creates and configures an Echo router.
// It initializes the session store, loads templates from the provided fs.FS,
// sets up logging and validation, and returns the Echo instance.
func New(tmplDir fs.FS, extraData map[string]interface{}, secret [64]byte) *echo.Echo {
	e := echo.New()

	//cookiePath := util.GetCookiePath()
	//cookieStore := sessions.NewCookieStore(secret[:32], secret[32:])
	//cookieStore.Options.Path = cookiePath
	//cookieStore.Options.HttpOnly = true
	//cookieStore.MaxAge(86400 * 7)

	cookieStore := sessions.NewCookieStore(secret[:32], secret[32:])
	cookieStore.Options.Path = util.GetCookiePath()
	cookieStore.Options.HttpOnly = true
	cookieStore.MaxAge(86400 * 7)

	e.Use(session.Middleware(cookieStore))

	// --- New middleware: Clear invalid session cookies ---
	// If session.Get fails (e.g. due to securecookie errors),
	// we clear the "session_token" cookie so that new sessions can be generated.
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if _, err := session.Get("session", c); err != nil {
				log.Debugf("session.Get failed: %v", err)
				// Clear the invalid session cookie.
				cookie := &http.Cookie{
					Name:     "session_token",
					Value:    "",
					Path:     util.GetCookiePath(),
					MaxAge:   -1,
					HttpOnly: true,
					SameSite: http.SameSiteLaxMode,
				}
				c.SetCookie(cookie)
			} else {
				log.Debug("Session retrieved successfully")
			}
			return next(c)
		}
	})
	// --- End new middleware ---

	// Load HTML template files as strings.
	tmplBaseString, err := util.StringFromEmbedFile(tmplDir, "base.html")
	if err != nil {
		log.Fatal(err)
	}
	tmplLoginString, err := util.StringFromEmbedFile(tmplDir, "login.html")
	if err != nil {
		log.Fatal(err)
	}
	tmplProfileString, err := util.StringFromEmbedFile(tmplDir, "profile.html")
	if err != nil {
		log.Fatal(err)
	}
	tmplClientsString, err := util.StringFromEmbedFile(tmplDir, "clients.html")
	if err != nil {
		log.Fatal(err)
	}
	tmplServerString, err := util.StringFromEmbedFile(tmplDir, "server.html")
	if err != nil {
		log.Fatal(err)
	}
	tmplGlobalSettingsString, err := util.StringFromEmbedFile(tmplDir, "global_settings.html")
	if err != nil {
		log.Fatal(err)
	}
	tmplUsersSettingsString, err := util.StringFromEmbedFile(tmplDir, "users_settings.html")
	if err != nil {
		log.Fatal(err)
	}
	tmplStatusString, err := util.StringFromEmbedFile(tmplDir, "status.html")
	if err != nil {
		log.Fatal(err)
	}
	tmplAPIKeysString, err := util.StringFromEmbedFile(tmplDir, "api_keys.html")
	if err != nil {
		log.Fatal(err)
	}
	tmplAPIStatisticsString, err := util.StringFromEmbedFile(tmplDir, "api_statistics.html")
	if err != nil {
		log.Fatal(err)
	}
	tmplSecuritySettingsString, err := util.StringFromEmbedFile(tmplDir, "security_settings.html")
	if err != nil {
		log.Fatal(err)
	}
	tmplSecurityStatisticsString, err := util.StringFromEmbedFile(tmplDir, "security_statistics.html")
	if err != nil {
		log.Fatal(err)
	}

	// Create a function map for templates.
	funcs := template.FuncMap{
		"StringsJoin": strings.Join,
		"tr": func(t interface{}, key string) string {
			// Helper function to access nested translation keys from the translation map
			// This receives the already-selected translation map for the current language
			if trans, ok := t.(i18n.Translation); ok {
				// Navigate through nested keys (e.g., "nav.new_client" -> nav -> new_client)
				keys := strings.Split(key, ".")
				var current interface{} = map[string]interface{}(trans)
				for _, k := range keys {
					if m, ok := current.(map[string]interface{}); ok {
						if val, exists := m[k]; exists {
							current = val
						} else {
							log.Debugf("Translation key not found: %s (missing key: %s)", key, k)
							return key // Key not found, return the key path as fallback
						}
					} else {
						log.Debugf("Translation structure error for key: %s (not a map at: %s)", key, k)
						return key // Not a map, return the key path as fallback
					}
				}
				if str, ok := current.(string); ok {
					return str
				}
				log.Debugf("Translation value is not a string for key: %s", key)
			} else {
				log.Debugf("Translation object is not valid for key: %s (type: %T)", key, t)
			}
			return key
		},
	}

	// Build the map of templates.
	templates := map[string]*template.Template{
		"login.html":               template.Must(template.New("login").Funcs(funcs).Parse(tmplLoginString)),
		"profile.html":             template.Must(template.New("profile").Funcs(funcs).Parse(tmplBaseString + tmplProfileString)),
		"clients.html":             template.Must(template.New("clients").Funcs(funcs).Parse(tmplBaseString + tmplClientsString)),
		"server.html":              template.Must(template.New("server").Funcs(funcs).Parse(tmplBaseString + tmplServerString)),
		"global_settings.html":     template.Must(template.New("global_settings").Funcs(funcs).Parse(tmplBaseString + tmplGlobalSettingsString)),
		"users_settings.html":      template.Must(template.New("users_settings").Funcs(funcs).Parse(tmplBaseString + tmplUsersSettingsString)),
		"status.html":              template.Must(template.New("status").Funcs(funcs).Parse(tmplBaseString + tmplStatusString)),
		"api_keys.html":            template.Must(template.New("api_keys").Funcs(funcs).Parse(tmplBaseString + tmplAPIKeysString)),
		"api_statistics.html":      template.Must(template.New("api_statistics").Funcs(funcs).Parse(tmplBaseString + tmplAPIStatisticsString)),
		"security_settings.html":   template.Must(template.New("security_settings").Funcs(funcs).Parse(tmplBaseString + tmplSecuritySettingsString)),
		"security_statistics.html": template.Must(template.New("security_statistics").Funcs(funcs).Parse(tmplBaseString + tmplSecurityStatisticsString)),
	}

	// Register GeoIP middleware
	handler.RegisterMiddlewares(e, "GeoLite2-City.mmdb")

	// Parse the log level from environment (default INFO).
	lvl, err := util.ParseLogLevel(util.LookupEnvOrString(util.LogLevel, "INFO"))
	if err != nil {
		log.Fatal(err)
	}

	// Configure the logger middleware.
	logConfig := middleware.DefaultLoggerConfig
	logConfig.Skipper = func(c echo.Context) bool {
		resp := c.Response()
		if resp.Status >= 500 && lvl > log.ERROR {
			return true
		} else if resp.Status >= 400 && lvl > log.WARN {
			return true
		} else if lvl > log.DEBUG {
			return true
		}
		return false
	}

	e.Logger.SetLevel(lvl)
	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.LoggerWithConfig(logConfig))
	e.HideBanner = true
	e.HidePort = lvl > log.INFO
	e.Validator = NewValidator() // Assume NewValidator is defined elsewhere.
	e.Renderer = &TemplateRegistry{
		templates: templates,
		extraData: extraData,
	}

	return e
}
