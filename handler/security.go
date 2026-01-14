package handler

import (
	"compress/gzip"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/oschwald/maxminddb-golang"
	"github.com/rs/xid"

	"github.com/swissmakers/wireguard-manager/model"
	"github.com/swissmakers/wireguard-manager/store"
	"github.com/swissmakers/wireguard-manager/util"
)

// SecurityMiddleware checks for IP blocks, brute force protection, and GeoIP rules
func SecurityMiddleware(db store.IStore) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip security checks for health endpoint
			if c.Path() == util.BasePath+"/_health" {
				return next(c)
			}

			ip := util.GetRealIP(c)

			// Get security settings
			settings, err := db.GetSecuritySettings()
			if err != nil {
				log.Warnf("Failed to get security settings: %v", err)
				// Continue without security checks if settings can't be loaded
				return next(c)
			}

			// Check IP blocking
			if settings.IPBlockingEnabled {
				blocked, err := db.IsIPBlocked(ip)
				if err != nil {
					log.Warnf("Failed to check IP block status: %v", err)
				} else if blocked {
					// Log security event
					event := model.SecurityEvent{
						ID:          xid.New().String(),
						EventType:   "blocked_ip",
						IP:          ip,
						Description: fmt.Sprintf("Blocked IP %s attempted to access %s", ip, c.Request().URL.Path),
						CreatedAt:   time.Now().UTC(),
					}
					_ = db.SaveSecurityEvent(event)

					return c.JSON(http.StatusForbidden, jsonHTTPResponse{
						Success: false,
						Message: "Access denied",
					})
				}
			}

			// Check brute force protection (only for login attempts)
			if settings.BruteForceEnabled && c.Path() == util.BasePath+"/login" && c.Request().Method == "POST" {
				attempt, err := db.GetBruteForceAttempt(ip)
				if err == nil && !attempt.BlockedUntil.IsZero() && time.Now().UTC().Before(attempt.BlockedUntil) {
					// Still blocked
					event := model.SecurityEvent{
						ID:          xid.New().String(),
						EventType:   "brute_force",
						IP:          ip,
						Description: fmt.Sprintf("Brute force blocked IP %s (blocked until %s)", ip, attempt.BlockedUntil.Format(time.RFC3339)),
						CreatedAt:   time.Now().UTC(),
					}
					_ = db.SaveSecurityEvent(event)

					return c.JSON(http.StatusTooManyRequests, jsonHTTPResponse{
						Success: false,
						Message: "Too many login attempts. Please try again later.",
					})
				}
			}

			return next(c)
		}
	}
}

// RecordFailedLogin records a failed login attempt for brute force protection
func RecordFailedLogin(db store.IStore, ip string, username string) {
	settings, err := db.GetSecuritySettings()
	if err != nil || !settings.BruteForceEnabled {
		return
	}

	// Get or create attempt record
	attempt, err := db.GetBruteForceAttempt(ip)
	if err != nil {
		// Create new attempt record
		attempt = model.BruteForceAttempt{
			IP:          ip,
			Attempts:    0,
			LastAttempt: time.Now().UTC(),
		}
	}

	// Check if attempt is within the time window
	windowStart := time.Now().UTC().Add(-time.Duration(settings.BruteForceWindowMinutes) * time.Minute)
	if attempt.LastAttempt.Before(windowStart) {
		// Reset attempts if outside window
		attempt.Attempts = 0
	}

	// Increment attempts
	attempt.Attempts++
	attempt.LastAttempt = time.Now().UTC()

	// Check if should be blocked
	if attempt.Attempts >= settings.BruteForceMaxAttempts {
		attempt.BlockedUntil = time.Now().UTC().Add(time.Duration(settings.BruteForceBlockMinutes) * time.Minute)

		// Log security event
		event := model.SecurityEvent{
			ID:          xid.New().String(),
			EventType:   "brute_force",
			IP:          ip,
			Username:    username,
			Description: fmt.Sprintf("Brute force detected from IP %s (user: %s) - blocked for %d minutes", ip, username, settings.BruteForceBlockMinutes),
			CreatedAt:   time.Now().UTC(),
		}
		_ = db.SaveSecurityEvent(event)

		log.Warnf("Brute force detected from IP %s - blocked for %d minutes", ip, settings.BruteForceBlockMinutes)
	}

	_ = db.SaveBruteForceAttempt(attempt)

	// Log failed login event
	event := model.SecurityEvent{
		ID:          xid.New().String(),
		EventType:   "failed_login",
		IP:          ip,
		Username:    username,
		Description: fmt.Sprintf("Failed login attempt from IP %s (user: %s) - attempt %d of %d", ip, username, attempt.Attempts, settings.BruteForceMaxAttempts),
		CreatedAt:   time.Now().UTC(),
	}
	_ = db.SaveSecurityEvent(event)
}

// ClearFailedLogins clears failed login attempts for an IP (called on successful login)
func ClearFailedLogins(db store.IStore, ip string) {
	_ = db.DeleteBruteForceAttempt(ip)
}

// SecuritySettingsPage renders the security settings admin page
func SecuritySettingsPage() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.Render(http.StatusOK, "security_settings.html", map[string]interface{}{
			"baseData": model.BaseData{
				Active:      "security-settings",
				CurrentUser: currentUser(c),
				Admin:       isAdmin(c),
			},
		})
	}
}

// SecurityStatisticsPage renders the security statistics admin page
func SecurityStatisticsPage() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.Render(http.StatusOK, "security_statistics.html", map[string]interface{}{
			"baseData": model.BaseData{
				Active:      "security-statistics",
				CurrentUser: currentUser(c),
				Admin:       isAdmin(c),
			},
		})
	}
}

// GetSecuritySettings returns the current security settings
func GetSecuritySettings(db store.IStore) echo.HandlerFunc {
	return func(c echo.Context) error {
		settings, err := db.GetSecuritySettings()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, jsonHTTPResponse{
				Success: false,
				Message: fmt.Sprintf("Cannot get security settings: %v", err),
			})
		}
		return c.JSON(http.StatusOK, settings)
	}
}

// UpdateSecuritySettings updates the security settings
func UpdateSecuritySettings(db store.IStore) echo.HandlerFunc {
	return func(c echo.Context) error {
		var settings model.SecuritySettings
		if err := c.Bind(&settings); err != nil {
			return c.JSON(http.StatusBadRequest, jsonHTTPResponse{
				Success: false,
				Message: "Invalid security settings data",
			})
		}

		settings.UpdatedAt = time.Now().UTC()
		if err := db.SaveSecuritySettings(settings); err != nil {
			return c.JSON(http.StatusInternalServerError, jsonHTTPResponse{
				Success: false,
				Message: fmt.Sprintf("Cannot update security settings: %v", err),
			})
		}

		log.Infof("Updated security settings")
		return c.JSON(http.StatusOK, jsonHTTPResponse{
			Success: true,
			Message: "Security settings updated successfully",
		})
	}
}

// GetSecurityEvents returns recent security events
func GetSecurityEvents(db store.IStore) echo.HandlerFunc {
	return func(c echo.Context) error {
		limit := 100
		events, err := db.GetSecurityEvents(limit)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, jsonHTTPResponse{
				Success: false,
				Message: fmt.Sprintf("Cannot get security events: %v", err),
			})
		}
		return c.JSON(http.StatusOK, events)
	}
}

// GetIPBlocks returns all IP blocks
func GetIPBlocks(db store.IStore) echo.HandlerFunc {
	return func(c echo.Context) error {
		blocks, err := db.GetIPBlocks()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, jsonHTTPResponse{
				Success: false,
				Message: fmt.Sprintf("Cannot get IP blocks: %v", err),
			})
		}
		return c.JSON(http.StatusOK, blocks)
	}
}

type createIPBlockRequest struct {
	IP        string `json:"ip"`
	Reason    string `json:"reason"`
	Permanent bool   `json:"permanent"`
	Hours     int    `json:"hours"` // Duration in hours if not permanent
}

// CreateIPBlock creates a new IP block
func CreateIPBlock(db store.IStore) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req createIPBlockRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, jsonHTTPResponse{
				Success: false,
				Message: "Invalid request data",
			})
		}

		if req.IP == "" {
			return c.JSON(http.StatusBadRequest, jsonHTTPResponse{
				Success: false,
				Message: "IP address is required",
			})
		}

		block := model.IPBlock{
			ID:        xid.New().String(),
			IP:        req.IP,
			Reason:    req.Reason,
			BlockedBy: currentUser(c),
			Permanent: req.Permanent,
			CreatedAt: time.Now().UTC(),
		}

		if !req.Permanent && req.Hours > 0 {
			block.ExpiresAt = time.Now().UTC().Add(time.Duration(req.Hours) * time.Hour)
		}

		if err := db.SaveIPBlock(block); err != nil {
			return c.JSON(http.StatusInternalServerError, jsonHTTPResponse{
				Success: false,
				Message: fmt.Sprintf("Cannot create IP block: %v", err),
			})
		}

		// Log security event
		event := model.SecurityEvent{
			ID:          xid.New().String(),
			EventType:   "blocked_ip",
			IP:          req.IP,
			Description: fmt.Sprintf("IP %s blocked by admin %s. Reason: %s", req.IP, currentUser(c), req.Reason),
			CreatedAt:   time.Now().UTC(),
		}
		_ = db.SaveSecurityEvent(event)

		log.Infof("IP %s blocked by admin %s", req.IP, currentUser(c))
		return c.JSON(http.StatusOK, jsonHTTPResponse{
			Success: true,
			Message: "IP blocked successfully",
		})
	}
}

type deleteIPBlockRequest struct {
	ID string `json:"id"`
}

// DeleteIPBlock removes an IP block
func DeleteIPBlock(db store.IStore) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req deleteIPBlockRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, jsonHTTPResponse{
				Success: false,
				Message: "Invalid request data",
			})
		}

		if err := db.DeleteIPBlock(req.ID); err != nil {
			return c.JSON(http.StatusInternalServerError, jsonHTTPResponse{
				Success: false,
				Message: fmt.Sprintf("Cannot delete IP block: %v", err),
			})
		}

		log.Infof("IP block removed by admin %s", currentUser(c))
		return c.JSON(http.StatusOK, jsonHTTPResponse{
			Success: true,
			Message: "IP block removed successfully",
		})
	}
}

// GetGeoIPRules returns all GeoIP rules
func GetGeoIPRules(db store.IStore) echo.HandlerFunc {
	return func(c echo.Context) error {
		rules, err := db.GetGeoIPRules()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, jsonHTTPResponse{
				Success: false,
				Message: fmt.Sprintf("Cannot get GeoIP rules: %v", err),
			})
		}
		return c.JSON(http.StatusOK, rules)
	}
}

type createGeoIPRuleRequest struct {
	CountryCode string `json:"country_code"`
	CountryName string `json:"country_name"`
	Action      string `json:"action"` // "block" or "allow"
}

// CreateGeoIPRule creates a new GeoIP rule
func CreateGeoIPRule(db store.IStore) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req createGeoIPRuleRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, jsonHTTPResponse{
				Success: false,
				Message: "Invalid request data",
			})
		}

		if req.CountryCode == "" || req.Action == "" {
			return c.JSON(http.StatusBadRequest, jsonHTTPResponse{
				Success: false,
				Message: "Country code and action are required",
			})
		}

		if req.Action != "block" && req.Action != "allow" {
			return c.JSON(http.StatusBadRequest, jsonHTTPResponse{
				Success: false,
				Message: "Action must be 'block' or 'allow'",
			})
		}

		rule := model.GeoIPRule{
			ID:          xid.New().String(),
			CountryCode: req.CountryCode,
			CountryName: req.CountryName,
			Action:      req.Action,
			CreatedBy:   currentUser(c),
			CreatedAt:   time.Now().UTC(),
		}

		if err := db.SaveGeoIPRule(rule); err != nil {
			return c.JSON(http.StatusInternalServerError, jsonHTTPResponse{
				Success: false,
				Message: fmt.Sprintf("Cannot create GeoIP rule: %v", err),
			})
		}

		log.Infof("GeoIP rule created by admin %s: %s -> %s", currentUser(c), req.CountryCode, req.Action)
		return c.JSON(http.StatusOK, jsonHTTPResponse{
			Success: true,
			Message: "GeoIP rule created successfully",
		})
	}
}

type deleteGeoIPRuleRequest struct {
	ID string `json:"id"`
}

// DeleteGeoIPRule removes a GeoIP rule
func DeleteGeoIPRule(db store.IStore) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req deleteGeoIPRuleRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, jsonHTTPResponse{
				Success: false,
				Message: "Invalid request data",
			})
		}

		if err := db.DeleteGeoIPRule(req.ID); err != nil {
			return c.JSON(http.StatusInternalServerError, jsonHTTPResponse{
				Success: false,
				Message: fmt.Sprintf("Cannot delete GeoIP rule: %v", err),
			})
		}

		log.Infof("GeoIP rule removed by admin %s", currentUser(c))
		return c.JSON(http.StatusOK, jsonHTTPResponse{
			Success: true,
			Message: "GeoIP rule removed successfully",
		})
	}
}

// GetSecurityStatistics returns aggregated security statistics
func GetSecurityStatistics(db store.IStore) echo.HandlerFunc {
	return func(c echo.Context) error {
		events, err := db.GetSecurityEvents(1000)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, jsonHTTPResponse{
				Success: false,
				Message: fmt.Sprintf("Cannot get security events: %v", err),
			})
		}

		// Aggregate statistics
		stats := map[string]interface{}{
			"total_events":       len(events),
			"failed_logins":      0,
			"blocked_ips":        0,
			"blocked_geoips":     0,
			"brute_force_blocks": 0,
			"top_ips":            make(map[string]int),
			"events_by_type":     make(map[string]int),
		}

		topIPs := make(map[string]int)
		eventsByType := make(map[string]int)

		for _, event := range events {
			eventsByType[event.EventType]++
			topIPs[event.IP]++

			switch event.EventType {
			case "failed_login":
				stats["failed_logins"] = stats["failed_logins"].(int) + 1
			case "blocked_ip":
				stats["blocked_ips"] = stats["blocked_ips"].(int) + 1
			case "blocked_geoip":
				stats["blocked_geoips"] = stats["blocked_geoips"].(int) + 1
			case "brute_force":
				stats["brute_force_blocks"] = stats["brute_force_blocks"].(int) + 1
			}
		}

		stats["top_ips"] = topIPs
		stats["events_by_type"] = eventsByType

		return c.JSON(http.StatusOK, stats)
	}
}

// Add GeoIP blocking functionality using maxminddb-golang
// Function to download GeoLite2-City.mmdb
func DownloadGeoLite2CityDB() error {
	url := "https://cdn.jsdelivr.net/npm/geolite2-city/GeoLite2-City.mmdb.gz"
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download GeoLite2-City.mmdb: %v", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			log.Warnf("Failed to close response body: %v", cerr)
		}
	}()

	// Save the file locally
	out, err := os.Create("GeoLite2-City.mmdb.gz")
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer func() {
		if cerr := out.Close(); cerr != nil {
			log.Warnf("Failed to close file: %v", cerr)
		}
	}()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save file: %v", err)
	}

	// Unzip the file
	return UnzipGeoLite2CityDB("GeoLite2-City.mmdb.gz", "GeoLite2-City.mmdb")
}

// Implement the UnzipGeoLite2CityDB function to handle .gz files
func UnzipGeoLite2CityDB(src, dest string) error {
	log.Infof("Unzipping %s to %s", src, dest)

	// Open the .gz file
	gzFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open .gz file: %w", err)
	}
	defer func() {
		if err := gzFile.Close(); err != nil {
			log.Warnf("Failed to close .gz file: %v", err)
		}
	}()

	// Create a gzip reader
	gzReader, err := gzip.NewReader(gzFile)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer func() {
		if err := gzReader.Close(); err != nil {
			log.Warnf("Failed to close gzip reader: %v", err)
		}
	}()

	// Create the destination file
	destFile, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer func() {
		if err := destFile.Close(); err != nil {
			log.Warnf("Failed to close destination file: %v", err)
		}
	}()

	// Copy the decompressed data to the destination file
	if _, err := io.Copy(destFile, gzReader); err != nil {
		return fmt.Errorf("failed to copy decompressed data: %w", err)
	}

	log.Infof("Successfully unzipped %s to %s", src, dest)

	// Delete the .gz file after successful extraction
	if err := os.Remove(src); err != nil {
		log.Warnf("Failed to delete .gz file %s: %v", src, err)
	} else {
		log.Infof("Successfully deleted .gz file %s", src)
	}

	return nil
}

// Ensure GeoIPMiddleware is used in the application
func RegisterMiddlewares(e *echo.Echo, dbPath string) {
	e.Use(GeoIPMiddleware(dbPath))
}

// Check if GeoLite2-City.mmdb exists, and download it if not
func EnsureGeoLite2CityDBExists() error {
	dbPath := "GeoLite2-City.mmdb"
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		log.Infof("GeoLite2-City.mmdb not found. Downloading...")
		if err := DownloadGeoLite2CityDB(); err != nil {
			return fmt.Errorf("failed to ensure GeoLite2-City.mmdb exists: %v", err)
		}
	}
	return nil
}

// Modify GeoIPMiddleware to ensure the database exists before activation
func GeoIPMiddleware(dbPath string) echo.MiddlewareFunc {
	if err := EnsureGeoLite2CityDBExists(); err != nil {
		log.Errorf("GeoIPMiddleware initialization failed: %v", err)
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				return c.JSON(http.StatusInternalServerError, map[string]string{"message": "GeoIP database unavailable."})
			}
		}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			mmdb, err := maxminddb.Open(dbPath)
			if err != nil {
				return fmt.Errorf("failed to open GeoLite2-City.mmdb: %v", err)
			}
			defer func() {
				if cerr := mmdb.Close(); cerr != nil {
					log.Warnf("Failed to close maxminddb: %v", cerr)
				}
			}()

			ip := net.ParseIP(util.GetRealIP(c))
			var record struct {
				Country struct {
					ISOCode string `maxminddb:"iso_code"`
				} `maxminddb:"country"`
			}
			if err := mmdb.Lookup(ip, &record); err != nil {
				return fmt.Errorf("failed to lookup IP: %v", err)
			}

			// Example: Block all requests from a specific country
			if record.Country.ISOCode == "CN" {
				return c.JSON(http.StatusForbidden, map[string]string{"message": "Access denied from your region."})
			}

			return next(c)
		}
	}
}

// NewGeoIPMiddleware checks if the client IP belongs to a blocked country
func NewGeoIPMiddleware(dbPath string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Extract client IP from the request
			clientIP := c.RealIP()

			// Open the GeoLite2 database
			db, err := maxminddb.Open(dbPath)
			if err != nil {
				log.Errorf("Failed to open GeoLite2 database: %v", err)
				return echo.NewHTTPError(http.StatusInternalServerError, "GeoIP database error")
			}
			defer db.Close()

			// Parse the client IP
			ip := net.ParseIP(clientIP)
			if ip == nil {
				log.Warnf("Invalid client IP: %s", clientIP)
				return echo.NewHTTPError(http.StatusBadRequest, "Invalid client IP")
			}

			// Query the GeoLite2 database for the country
			var record struct {
				Country struct {
					ISOCode string `maxminddb:"iso_code"`
				} `maxminddb:"country"`
			}
			if err := db.Lookup(ip, &record); err != nil {
				log.Errorf("Failed to lookup IP in GeoLite2 database: %v", err)
				return echo.NewHTTPError(http.StatusInternalServerError, "GeoIP lookup error")
			}

			// Check if the country is blocked
			blockedCountries := []string{"CN", "RU", "IR"} // Example blocked countries
			for _, blocked := range blockedCountries {
				if record.Country.ISOCode == blocked {
					log.Infof("Blocked client IP %s from country %s", clientIP, record.Country.ISOCode)
					return echo.NewHTTPError(http.StatusForbidden, "Access from your country is restricted")
				}
			}

			// Proceed to the next handler
			return next(c)
		}
	}
}

// Load the GeoLite2-City.mmdb database for use
func LoadGeoLite2CityDB(dbPath string) (*maxminddb.Reader, error) {
	log.Infof("Loading GeoLite2-City.mmdb from %s", dbPath)
	mmdb, err := maxminddb.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load GeoLite2-City.mmdb: %w", err)
	}
	// Ensure the resource is closed properly
	defer func() {
		if cerr := mmdb.Close(); cerr != nil {
			log.Warnf("Failed to close GeoLite2-City.mmdb: %v", cerr)
		}
	}()
	log.Infof("Successfully loaded GeoLite2-City.mmdb")
	return mmdb, nil
}
