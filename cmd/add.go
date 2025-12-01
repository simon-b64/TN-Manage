package cmd

import (
	"fmt"
	"strconv"

	"github.com/nox/tnmanage/pkg/truenas"
	"github.com/spf13/cobra"
)

var (
	addServer string
	addToken  string
	nfsHosts  []string
)

var addCmd = &cobra.Command{
	Use:   "add <poolname> <datasetname> <max-size-gb>",
	Short: "Add a new dataset",
	Long:  `Add a new dataset to TrueNAS with optional NFS share`,
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		poolName := args[0]
		datasetName := args[1]
		maxSizeGB, err := strconv.Atoi(args[2])
		if err != nil {
			return fmt.Errorf("invalid max size: %w", err)
		}

		var client *truenas.Client

		if addServer != "" && addToken != "" {
			client, err = truenas.NewClientWithParams(addServer, addToken)
		} else {
			client, err = truenas.NewClient()
		}

		if err != nil {
			return fmt.Errorf("failed to create TrueNAS client: %w", err)
		}

		// Create the dataset
		datasetID, err := client.CreateDataset(poolName, datasetName, maxSizeGB)
		if err != nil {
			return fmt.Errorf("failed to create dataset: %w", err)
		}

		fmt.Printf("Successfully created dataset '%s'\n", datasetID)

		// Create NFS share if --nfs flag is provided
		if len(nfsHosts) > 0 {
			share := &truenas.NFSShare{
				Path:         fmt.Sprintf("/mnt/%s", datasetID),
				Comment:      datasetName,
				Hosts:        nfsHosts,
				MapRootUser:  "root",
				MapRootGroup: "wheel",
				ReadOnly:     false,
			}

			nfsID, err := client.CreateNFSShare(share)
			if err != nil {
				return fmt.Errorf("failed to create NFS share: %w", err)
			}

			fmt.Printf("Successfully created NFS share (ID: %d) for hosts: %v\n", nfsID, nfsHosts)
		}

		return nil
	},
}

func init() {
	addCmd.Flags().StringVar(&addServer, "server", "", "TrueNAS server URL (e.g., https://192.168.1.100)")
	addCmd.Flags().StringVar(&addToken, "token", "", "TrueNAS API token")
	addCmd.Flags().StringSliceVar(&nfsHosts, "nfs", []string{}, "Authorized hosts for NFS share (creates NFS share if specified)")
}
