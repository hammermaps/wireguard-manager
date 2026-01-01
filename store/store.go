package store

import (
	"github.com/swissmakers/wireguard-manager/model"
)

// IStore defines the interface for data storage used in the application.
// It abstracts the methods for user management, server configuration,
// client management, and hash tracking.
type IStore interface {
	// Initialization
	Init() error

	// User Management
	GetUsers() ([]model.User, error)
	GetUserByName(username string) (model.User, error)
	SaveUser(user model.User) error
	DeleteUser(username string) error

	// Global Settings and Server Configuration
	GetGlobalSettings() (model.GlobalSetting, error)
	GetServer() (model.Server, error)
	SaveServerInterface(serverInterface model.ServerInterface) error
	SaveServerKeyPair(serverKeyPair model.ServerKeypair) error
	SaveGlobalSettings(globalSettings model.GlobalSetting) error

	// Client Management
	GetClients(hasQRCode bool) ([]model.ClientData, error)
	GetClientByID(clientID string, qrCode model.QRCodeSettings) (model.ClientData, error)
	SaveClient(client model.Client) error
	DeleteClient(clientID string) error

	// File Storage Path
	GetPath() string

	// Hash Management for Config Change Detection
	SaveHashes(hashes model.ClientServerHashes) error
	GetHashes() (model.ClientServerHashes, error)

	// API Key Management
	GetAPIKeys() ([]model.APIKey, error)
	GetAPIKeyByID(keyID string) (model.APIKey, error)
	GetAPIKeyByKey(key string) (model.APIKey, error)
	SaveAPIKey(key model.APIKey) error
	DeleteAPIKey(keyID string) error

	// API Access Log Management
	SaveAPIAccessLog(log model.APIAccessLog) error
	GetAPIAccessLogs(limit int) ([]model.APIAccessLog, error)
	GetAPIAccessLogsByKeyID(keyID string, limit int) ([]model.APIAccessLog, error)

	// Security Management
	GetSecuritySettings() (model.SecuritySettings, error)
	SaveSecuritySettings(settings model.SecuritySettings) error

	// Security Events
	SaveSecurityEvent(event model.SecurityEvent) error
	GetSecurityEvents(limit int) ([]model.SecurityEvent, error)
	GetSecurityEventsByType(eventType string, limit int) ([]model.SecurityEvent, error)

	// IP Blocking
	GetIPBlocks() ([]model.IPBlock, error)
	GetIPBlockByIP(ip string) (model.IPBlock, error)
	SaveIPBlock(block model.IPBlock) error
	DeleteIPBlock(id string) error
	IsIPBlocked(ip string) (bool, error)

	// GeoIP Rules
	GetGeoIPRules() ([]model.GeoIPRule, error)
	GetGeoIPRuleByCountry(countryCode string) (model.GeoIPRule, error)
	SaveGeoIPRule(rule model.GeoIPRule) error
	DeleteGeoIPRule(id string) error

	// Brute Force Protection
	GetBruteForceAttempt(ip string) (model.BruteForceAttempt, error)
	SaveBruteForceAttempt(attempt model.BruteForceAttempt) error
	DeleteBruteForceAttempt(ip string) error
	CleanupExpiredBruteForceAttempts() error
}
