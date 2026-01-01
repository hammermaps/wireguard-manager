package util

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/labstack/gommon/log"
	"github.com/swissmakers/wireguard-manager/model"
)

// GetWireGuardInterface extracts the interface name from the config file path
// e.g., "/etc/wireguard/wg0.conf" -> "wg0"
func GetWireGuardInterface(configPath string) string {
	if configPath == "" {
		log.Warn("Empty config path provided to GetWireGuardInterface, using default 'wg0'")
		return "wg0"
	}

	// Extract filename from path using filepath.Base
	filename := filepath.Base(configPath)
	// Remove .conf extension
	interfaceName := strings.TrimSuffix(filename, ".conf")

	if interfaceName == "" {
		log.Warn("Could not extract interface name from config path, using default 'wg0'")
		return "wg0"
	}

	return interfaceName
}

// ReloadWireGuard reloads the WireGuard configuration using wg syncconf
// This applies changes without disrupting existing connections
func ReloadWireGuard(settings model.GlobalSetting) error {
	interfaceName := GetWireGuardInterface(settings.ConfigFilePath)

	// First, strip the wg-quick specific directives from the config
	stripCmd := exec.Command("wg-quick", "strip", interfaceName)
	stripOutput, err := stripCmd.Output()
	if err != nil {
		log.Errorf("Failed to strip config for interface %s: %v", interfaceName, err)
		return fmt.Errorf("failed to strip config: %w", err)
	}

	// Now apply the stripped config using wg syncconf
	syncCmd := exec.Command("wg", "syncconf", interfaceName, "/dev/stdin")
	syncCmd.Stdin = strings.NewReader(string(stripOutput))

	output, err := syncCmd.CombinedOutput()
	if err != nil {
		log.Errorf("Failed to reload WireGuard interface %s: %v, output: %s", interfaceName, err, string(output))
		return fmt.Errorf("failed to reload WireGuard: %w", err)
	}

	log.Infof("Successfully reloaded WireGuard interface %s", interfaceName)
	return nil
}

// StartWireGuard starts the WireGuard interface
func StartWireGuard(settings model.GlobalSetting) error {
	interfaceName := GetWireGuardInterface(settings.ConfigFilePath)

	cmd := exec.Command("wg-quick", "up", interfaceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf("Failed to start WireGuard interface %s: %v, output: %s", interfaceName, err, string(output))
		return fmt.Errorf("failed to start WireGuard: %w", err)
	}

	log.Infof("Successfully started WireGuard interface %s", interfaceName)
	return nil
}

// StopWireGuard stops the WireGuard interface
func StopWireGuard(settings model.GlobalSetting) error {
	interfaceName := GetWireGuardInterface(settings.ConfigFilePath)

	cmd := exec.Command("wg-quick", "down", interfaceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf("Failed to stop WireGuard interface %s: %v, output: %s", interfaceName, err, string(output))
		return fmt.Errorf("failed to stop WireGuard: %w", err)
	}

	log.Infof("Successfully stopped WireGuard interface %s", interfaceName)
	return nil
}

// RestartWireGuard restarts the WireGuard interface
func RestartWireGuard(settings model.GlobalSetting) error {
	// First stop the interface
	if err := StopWireGuard(settings); err != nil {
		// Log the error but continue - interface might not be running
		log.Warnf("Failed to stop WireGuard during restart: %v", err)
	}

	// Then start it
	return StartWireGuard(settings)
}

// GetWireGuardStatus checks if the WireGuard interface is running
func GetWireGuardStatus(settings model.GlobalSetting) (bool, error) {
	interfaceName := GetWireGuardInterface(settings.ConfigFilePath)

	cmd := exec.Command("wg", "show", interfaceName)
	err := cmd.Run()
	if err != nil {
		// If the command fails, the interface is likely not running
		return false, nil
	}

	return true, nil
}
