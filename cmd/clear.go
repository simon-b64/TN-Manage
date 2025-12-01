package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/nox/tnmanage/pkg/truenas"
	"github.com/spf13/cobra"
)

var (
	clearServer string
	clearToken  string
	clearForce  bool
)

var clearCmd = &cobra.Command{
	Use:   "clear <datasetname>",
	Short: "Wipe/clear all data from a dataset",
	Long:  `Delete all contents of a dataset. This operation cannot be undone!`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		datasetName := args[0]

		// Confirmation prompt unless --force is used
		if !clearForce {
			fmt.Printf("WARNING: This will DELETE ALL DATA in dataset '%s'\n", datasetName)
			fmt.Print("Are you sure? (y/n): ")

			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read confirmation: %w", err)
			}

			response = strings.TrimSpace(strings.ToLower(response))
			if response != "y" && response != "yes" {
				fmt.Println("Operation cancelled")
				return nil
			}
		}

		var client *truenas.Client
		var err error

		if clearServer != "" && clearToken != "" {
			client, err = truenas.NewClientWithParams(clearServer, clearToken)
		} else {
			client, err = truenas.NewClient()
		}

		if err != nil {
			return fmt.Errorf("failed to create TrueNAS client: %w", err)
		}

		if err := client.ClearDataset(datasetName); err != nil {
			return fmt.Errorf("failed to clear dataset: %w", err)
		}

		fmt.Printf("Successfully cleared dataset '%s'\n", datasetName)
		return nil
	},
}

func init() {
	clearCmd.Flags().StringVar(&clearServer, "server", "", "TrueNAS server URL (e.g., https://192.168.1.100)")
	clearCmd.Flags().StringVar(&clearToken, "token", "", "TrueNAS API token")
	clearCmd.Flags().BoolVarP(&clearForce, "force", "f", false, "Skip confirmation prompt")
}
