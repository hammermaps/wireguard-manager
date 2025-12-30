package handler

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
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

// Request/Response structures for API key management
type createAPIKeyRequest struct {
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

type createAPIKeyResponse struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	APIKey  *model.APIKey  `json:"api_key,omitempty"`
	RawKey  string         `json:"raw_key,omitempty"`
}

type updateAPIKeyRequest struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
	Enabled     bool     `json:"enabled"`
}

type deleteAPIKeyRequest struct {
	ID string `json:"id"`
}

type setGroupStatusRequest struct {
	Group   string `json:"group"`
	Enabled bool   `json:"enabled"`
}

// generateAPIKey generates a random API key
func generateAPIKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// hashAPIKey hashes an API key for storage
func hashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// APIKeyManagementPage renders the API key management page
func APIKeyManagementPage(db store.IStore) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.Render(http.StatusOK, "api_keys.html", map[string]interface{}{})
	}
}

// GetAPIKeys returns all API keys
func GetAPIKeys(db store.IStore) echo.HandlerFunc {
	return func(c echo.Context) error {
		keys, err := db.GetAPIKeys()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, jsonHTTPResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to retrieve API keys: %v", err),
			})
		}

		// Remove sensitive key data before sending to client
		for i := range keys {
			keys[i].Key = ""
		}

		return c.JSON(http.StatusOK, keys)
	}
}

// CreateAPIKey creates a new API key
func CreateAPIKey(db store.IStore) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req createAPIKeyRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, createAPIKeyResponse{
				Success: false,
				Message: "Invalid request data",
			})
		}

		if req.Name == "" {
			return c.JSON(http.StatusBadRequest, createAPIKeyResponse{
				Success: false,
				Message: "API key name is required",
			})
		}

		// Generate a new API key
		rawKey, err := generateAPIKey()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, createAPIKeyResponse{
				Success: false,
				Message: "Failed to generate API key",
			})
		}

		// Create API key object
		now := time.Now().UTC()
		apiKey := model.APIKey{
			ID:          xid.New().String(),
			Name:        req.Name,
			Key:         hashAPIKey(rawKey),
			KeyPrefix:   rawKey[:8],
			Permissions: req.Permissions,
			Enabled:     true,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		// Save to database
		if err := db.SaveAPIKey(apiKey); err != nil {
			return c.JSON(http.StatusInternalServerError, createAPIKeyResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to save API key: %v", err),
			})
		}

		log.Infof("Created API key: %s (ID: %s)", apiKey.Name, apiKey.ID)

		// Return the API key with the raw key (only shown once)
		apiKey.Key = ""
		return c.JSON(http.StatusOK, createAPIKeyResponse{
			Success: true,
			Message: "API key created successfully",
			APIKey:  &apiKey,
			RawKey:  rawKey,
		})
	}
}

// UpdateAPIKey updates an existing API key
func UpdateAPIKey(db store.IStore) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req updateAPIKeyRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, jsonHTTPResponse{
				Success: false,
				Message: "Invalid request data",
			})
		}

		// Get existing API key
		apiKey, err := db.GetAPIKeyByID(req.ID)
		if err != nil {
			return c.JSON(http.StatusNotFound, jsonHTTPResponse{
				Success: false,
				Message: "API key not found",
			})
		}

		// Update fields
		apiKey.Name = req.Name
		apiKey.Permissions = req.Permissions
		apiKey.Enabled = req.Enabled
		apiKey.UpdatedAt = time.Now().UTC()

		// Save to database
		if err := db.SaveAPIKey(apiKey); err != nil {
			return c.JSON(http.StatusInternalServerError, jsonHTTPResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to update API key: %v", err),
			})
		}

		log.Infof("Updated API key: %s (ID: %s)", apiKey.Name, apiKey.ID)

		return c.JSON(http.StatusOK, jsonHTTPResponse{
			Success: true,
			Message: "API key updated successfully",
		})
	}
}

// DeleteAPIKey deletes an API key
func DeleteAPIKey(db store.IStore) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req deleteAPIKeyRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, jsonHTTPResponse{
				Success: false,
				Message: "Invalid request data",
			})
		}

		// Delete from database
		if err := db.DeleteAPIKey(req.ID); err != nil {
			return c.JSON(http.StatusInternalServerError, jsonHTTPResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to delete API key: %v", err),
			})
		}

		log.Infof("Deleted API key with ID: %s", req.ID)

		return c.JSON(http.StatusOK, jsonHTTPResponse{
			Success: true,
			Message: "API key deleted successfully",
		})
	}
}

// SetGroupStatus enables or disables all clients in a group
func SetGroupStatus(db store.IStore) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req setGroupStatusRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, jsonHTTPResponse{
				Success: false,
				Message: "Invalid request data",
			})
		}

		if req.Group == "" {
			return c.JSON(http.StatusBadRequest, jsonHTTPResponse{
				Success: false,
				Message: "Group name is required",
			})
		}

		// Get all clients
		clients, err := db.GetClients(false)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, jsonHTTPResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to retrieve clients: %v", err),
			})
		}

		// Update clients in the specified group
		updatedCount := 0
		for _, clientData := range clients {
			if clientData.Client.Group == req.Group {
				client := *clientData.Client
				client.Enabled = req.Enabled
				if err := db.SaveClient(client); err != nil {
					log.Errorf("Failed to update client %s: %v", client.ID, err)
					continue
				}
				updatedCount++
			}
		}

		status := "enabled"
		if !req.Enabled {
			status = "disabled"
		}

		log.Infof("Changed status of %d clients in group '%s' to %s", updatedCount, req.Group, status)

		return c.JSON(http.StatusOK, jsonHTTPResponse{
			Success: true,
			Message: fmt.Sprintf("Successfully %s %d client(s) in group '%s'", status, updatedCount, req.Group),
		})
	}
}

// APIStatisticsPage renders the API statistics page
func APIStatisticsPage(db store.IStore) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.Render(http.StatusOK, "api_statistics.html", map[string]interface{}{})
	}
}

// GetAPIStatistics returns API usage statistics
func GetAPIStatistics(db store.IStore) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get recent access logs (last 1000)
		logs, err := db.GetAPIAccessLogs(1000)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, jsonHTTPResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to retrieve API logs: %v", err),
			})
		}

		// Get all API keys for reference
		keys, err := db.GetAPIKeys()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, jsonHTTPResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to retrieve API keys: %v", err),
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"logs":      logs,
			"api_keys":  keys,
			"log_count": len(logs),
		})
	}
}

// ValidateAPIKey middleware validates API key authentication
func ValidateAPIKey(db store.IStore) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get API key from Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, jsonHTTPResponse{
					Success: false,
					Message: "Missing Authorization header",
				})
			}

			// Expected format: "Bearer <api_key>"
			var apiKey string
			if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				apiKey = authHeader[7:]
			} else {
				return c.JSON(http.StatusUnauthorized, jsonHTTPResponse{
					Success: false,
					Message: "Invalid Authorization header format",
				})
			}

			// Hash the provided key and look it up
			hashedKey := hashAPIKey(apiKey)
			keyData, err := db.GetAPIKeyByKey(hashedKey)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, jsonHTTPResponse{
					Success: false,
					Message: "Invalid API key",
				})
			}

			// Check if key is enabled
			if !keyData.Enabled {
				return c.JSON(http.StatusForbidden, jsonHTTPResponse{
					Success: false,
					Message: "API key is disabled",
				})
			}

			// Update last used timestamp
			now := time.Now().UTC()
			keyData.LastUsedAt = &now
			db.SaveAPIKey(keyData)

			// Log the API access
			logEntry := model.APIAccessLog{
				ID:         xid.New().String(),
				APIKeyID:   keyData.ID,
				APIKeyName: keyData.Name,
				Endpoint:   c.Request().URL.Path,
				Method:     c.Request().Method,
				IPAddress:  util.GetClientIPFromRequest(c.Request().RemoteAddr, c.Request().Header.Get("X-Forwarded-For"), c.Request().Header.Get("X-Real-IP")),
				UserAgent:  c.Request().UserAgent(),
				StatusCode: 0, // Will be updated after request
				Timestamp:  now,
			}
			db.SaveAPIAccessLog(logEntry)

			// Store API key info in context for later use
			c.Set("api_key", keyData)

			return next(c)
		}
	}
}

// CheckAPIPermission middleware checks if the API key has the required permission
func CheckAPIPermission(permission string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			keyData, ok := c.Get("api_key").(model.APIKey)
			if !ok {
				return c.JSON(http.StatusForbidden, jsonHTTPResponse{
					Success: false,
					Message: "API key context not found",
				})
			}

			// Check if key has the required permission
			hasPermission := false
			for _, perm := range keyData.Permissions {
				if perm == permission {
					hasPermission = true
					break
				}
			}

			if !hasPermission {
				return c.JSON(http.StatusForbidden, jsonHTTPResponse{
					Success: false,
					Message: fmt.Sprintf("API key does not have required permission: %s", permission),
				})
			}

			return next(c)
		}
	}
}
