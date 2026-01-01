package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
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
