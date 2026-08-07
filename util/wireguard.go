package util

import (
	"bufio"
	"fmt"
	"os"
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

	// Strip wg-quick-specific directives from the config file directly,
	// avoiding the need to call `wg-quick strip` which requires root/sudo.
	stripped, err := stripWGQuickConfig(settings.ConfigFilePath)
	if err != nil {
		log.Errorf("Failed to strip config from %s: %v", settings.ConfigFilePath, err)
		return fmt.Errorf("failed to strip config: %w", err)
	}

	// Apply the stripped config using wg syncconf
	syncCmd := exec.Command("wg", "syncconf", interfaceName, "/dev/stdin")
	syncCmd.Stdin = strings.NewReader(stripped)

	output, err := syncCmd.CombinedOutput()
	if err != nil {
		log.Errorf("Failed to reload WireGuard interface %s: %v, output: %s", interfaceName, err, string(output))
		return fmt.Errorf("failed to reload WireGuard: %w", err)
	}

	log.Infof("Successfully reloaded WireGuard interface %s", interfaceName)
	return nil
}

// stripWGQuickConfig reads a wg-quick config file and removes wg-quick-specific
// directives, returning only the lines that `wg syncconf` understands.
// This replicates the behavior of `wg-quick strip` without requiring root.
func stripWGQuickConfig(configPath string) (string, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return "", fmt.Errorf("cannot open config file %s: %w", configPath, err)
	}
	defer file.Close()

	// Keys that are wg-quick-specific and should be stripped.
	wgQuickKeys := map[string]bool{
		"Address":     true,
		"DNS":         true,
		"MTU":         true,
		"Table":       true,
		"PreUp":       true,
		"PostUp":      true,
		"PreDown":     true,
		"PostDown":    true,
		"SaveConfig":  true,
	}

	var result strings.Builder
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		stripped := strings.TrimSpace(line)

		// Keep comments, empty lines, and section headers.
		if stripped == "" || strings.HasPrefix(stripped, "#") || strings.HasPrefix(stripped, "[") {
			result.WriteString(line)
			result.WriteString("\n")
			continue
		}

		// Extract the key (everything before the first '=').
		key := strings.TrimSpace(strings.SplitN(stripped, "=", 2)[0])
		if !wgQuickKeys[key] {
			result.WriteString(line)
			result.WriteString("\n")
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading config file: %w", err)
	}

	return result.String(), nil
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
