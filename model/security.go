package model

import "time"

// SecurityEvent represents a security-related event (login attempt, blocked request, etc.)
type SecurityEvent struct {
	ID          string    `json:"id"`
	EventType   string    `json:"event_type"` // "failed_login", "blocked_ip", "blocked_geoip", "brute_force"
	IP          string    `json:"ip"`
	Country     string    `json:"country,omitempty"`
	Username    string    `json:"username,omitempty"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// IPBlock represents a blocked IP address
type IPBlock struct {
	ID          string    `json:"id"`
	IP          string    `json:"ip"`
	Reason      string    `json:"reason"`
	BlockedBy   string    `json:"blocked_by"` // username of admin who blocked it
	Permanent   bool      `json:"permanent"`
	ExpiresAt   time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// GeoIPRule represents a country-based blocking rule
type GeoIPRule struct {
	ID          string    `json:"id"`
	CountryCode string    `json:"country_code"` // ISO 3166-1 alpha-2 country code
	CountryName string    `json:"country_name"`
	Action      string    `json:"action"` // "block" or "allow"
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}

// BruteForceAttempt tracks login attempts from an IP
type BruteForceAttempt struct {
	IP           string    `json:"ip"`
	Attempts     int       `json:"attempts"`
	LastAttempt  time.Time `json:"last_attempt"`
	BlockedUntil time.Time `json:"blocked_until,omitempty"`
}

// SecuritySettings represents global security configuration
type SecuritySettings struct {
	// Brute Force Protection
	BruteForceEnabled       bool `json:"brute_force_enabled"`
	BruteForceMaxAttempts   int  `json:"brute_force_max_attempts"`   // Max failed login attempts
	BruteForceWindowMinutes int  `json:"brute_force_window_minutes"` // Time window for attempts
	BruteForceBlockMinutes  int  `json:"brute_force_block_minutes"`  // Block duration

	// IP Blocking
	IPBlockingEnabled bool `json:"ip_blocking_enabled"`

	// GeoIP Blocking
	GeoIPEnabled       bool   `json:"geoip_enabled"`
	GeoIPDefaultAction string `json:"geoip_default_action"` // "allow" or "block"

	UpdatedAt time.Time `json:"updated_at"`
}

// DefaultSecuritySettings returns default security settings
func DefaultSecuritySettings() SecuritySettings {
	return SecuritySettings{
		BruteForceEnabled:       true,
		BruteForceMaxAttempts:   5,
		BruteForceWindowMinutes: 15,
		BruteForceBlockMinutes:  30,
		IPBlockingEnabled:       true,
		GeoIPEnabled:            false,
		GeoIPDefaultAction:      "allow",
		UpdatedAt:               time.Now().UTC(),
	}
}
