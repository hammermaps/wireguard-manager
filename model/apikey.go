package model

import (
	"time"
)

// APIKey represents an API key for external access
type APIKey struct {
	// ID is a unique identifier for the API key
	ID string `json:"id"`

	// Name is a friendly name for the API key
	Name string `json:"name"`

	// Key is the actual API key value (hashed in storage)
	Key string `json:"key"`

	// KeyPrefix is the first 8 characters of the key for identification
	KeyPrefix string `json:"key_prefix"`

	// Permissions defines what actions this key can perform
	Permissions []string `json:"permissions"`

	// Enabled indicates if the key is currently active
	Enabled bool `json:"enabled"`

	// CreatedAt is the timestamp when the key was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is the timestamp of the key's last update
	UpdatedAt time.Time `json:"updated_at"`

	// LastUsedAt is the timestamp of the key's last use
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

// APIKeyPermission defines available permissions for API keys
const (
	PermissionReadClients   = "read:clients"
	PermissionWriteClients  = "write:clients"
	PermissionReadServer    = "read:server"
	PermissionWriteServer   = "write:server"
	PermissionManageGroups  = "manage:groups"
	PermissionReadStats     = "read:stats"
)

// APIAccessLog represents a log entry for API access
type APIAccessLog struct {
	// ID is a unique identifier for the log entry
	ID string `json:"id"`

	// APIKeyID is the ID of the API key that was used
	APIKeyID string `json:"api_key_id"`

	// APIKeyName is the name of the API key for display
	APIKeyName string `json:"api_key_name"`

	// Endpoint is the API endpoint that was accessed
	Endpoint string `json:"endpoint"`

	// Method is the HTTP method used
	Method string `json:"method"`

	// IPAddress is the client's IP address
	IPAddress string `json:"ip_address"`

	// UserAgent is the client's user agent
	UserAgent string `json:"user_agent"`

	// StatusCode is the HTTP response status code
	StatusCode int `json:"status_code"`

	// Timestamp is when the access occurred
	Timestamp time.Time `json:"timestamp"`
}
