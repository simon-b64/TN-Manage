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
	removeServer string
	removeToken  string
	removeForce  bool
)

var removeCmd = &cobra.Command{
	Use:   "remove <datasetname>",
	Short: "Remove a dataset",
	Long:  `Remove a dataset from TrueNAS by name`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		datasetName := args[0]

		// Confirmation prompt unless --force is used
		if !removeForce {
			fmt.Printf("WARNING: This will permanently DELETE dataset '%s' and all its data\n", datasetName)
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

		if removeServer != "" && removeToken != "" {
			client, err = truenas.NewClientWithParams(removeServer, removeToken)
		} else {
			client, err = truenas.NewClient()
		}

		if err != nil {
			return fmt.Errorf("failed to create TrueNAS client: %w", err)
		}

		if err := client.DeleteDataset(datasetName); err != nil {
			return fmt.Errorf("failed to delete dataset: %w", err)
		}

		fmt.Printf("Successfully removed dataset '%s'\n", datasetName)
		return nil
	},
}

func init() {
	removeCmd.Flags().StringVar(&removeServer, "server", "", "TrueNAS server URL (e.g., https://192.168.1.100)")
	removeCmd.Flags().StringVar(&removeToken, "token", "", "TrueNAS API token")
	removeCmd.Flags().BoolVarP(&removeForce, "force", "f", false, "Skip confirmation prompt")
}
