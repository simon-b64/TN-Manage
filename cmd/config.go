package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure TrueNAS connection settings",
	Long:  `Configure TrueNAS server URL and API token`,
}

var configServerCmd = &cobra.Command{
	Use:   "server <serverurl>",
	Short: "Set the TrueNAS server URL",
	Long:  `Set the TrueNAS server URL and save it to configuration`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL := args[0]

		if err := saveConfig("TRUENAS_URL", serverURL); err != nil {
			return fmt.Errorf("failed to save server URL: %w", err)
		}

		fmt.Printf("Server URL set to: %s\n", serverURL)
		fmt.Println("Configuration saved to ~/.tnmanage")
		return nil
	},
}

var configTokenCmd = &cobra.Command{
	Use:   "token <token>",
	Short: "Set the TrueNAS API token",
	Long:  `Set the TrueNAS API token and save it to configuration`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		token := args[0]

		if err := saveConfig("TRUENAS_API_KEY", token); err != nil {
			return fmt.Errorf("failed to save API token: %w", err)
		}

		fmt.Println("API token saved successfully")
		fmt.Println("Configuration saved to ~/.tnmanage")
		return nil
	},
}

func init() {
	configCmd.AddCommand(configServerCmd)
	configCmd.AddCommand(configTokenCmd)
}

func getConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".tnmanage"), nil
}

func saveConfig(key, value string) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	// Read existing config if it exists
	config := make(map[string]string)
	if data, err := os.ReadFile(configPath); err == nil {
		// Parse existing config
		lines := string(data)
		for _, line := range splitLines(lines) {
			if len(line) > 0 && line[0] != '#' {
				parts := splitKeyValue(line)
				if len(parts) == 2 {
					config[parts[0]] = parts[1]
				}
			}
		}
	}

	// Update the config
	config[key] = value

	// Write config file
	var content string
	content += "# TrueNAS Configuration\n"
	content += "# This file is automatically managed by tnmanage\n\n"
	if url, ok := config["TRUENAS_URL"]; ok {
		content += fmt.Sprintf("TRUENAS_URL=%s\n", url)
	}
	if token, ok := config["TRUENAS_API_KEY"]; ok {
		content += fmt.Sprintf("TRUENAS_API_KEY=%s\n", token)
	}

	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		return err
	}

	return nil
}

func splitLines(s string) []string {
	var lines []string
	var line string
	for _, c := range s {
		if c == '\n' {
			lines = append(lines, line)
			line = ""
		} else {
			line += string(c)
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines
}

func splitKeyValue(s string) []string {
	for i, c := range s {
		if c == '=' {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}

// LoadConfig loads configuration from file and sets environment variables
func LoadConfig() error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		// Config file doesn't exist, that's ok
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	// Parse and set environment variables
	lines := splitLines(string(data))
	for _, line := range lines {
		if len(line) > 0 && line[0] != '#' {
			parts := splitKeyValue(line)
			if len(parts) == 2 {
				// Only set if not already set in environment
				if os.Getenv(parts[0]) == "" {
					os.Setenv(parts[0], parts[1])
				}
			}
		}
	}

	return nil
}
