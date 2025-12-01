package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/nox/tnmanage/pkg/truenas"
	"github.com/spf13/cobra"
)

var (
	server string
	token  string
)

var listCmd = &cobra.Command{
	Use:   "list <poolname>",
	Short: "List all datasets in a pool",
	Long:  `List all datasets in a specified pool on TrueNAS`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		poolName := args[0]

		var client *truenas.Client
		var err error

		if server != "" && token != "" {
			client, err = truenas.NewClientWithParams(server, token)
		} else {
			client, err = truenas.NewClient()
		}

		if err != nil {
			return fmt.Errorf("failed to create TrueNAS client: %w", err)
		}

		datasets, err := client.ListDatasets(poolName)
		if err != nil {
			return fmt.Errorf("failed to list datasets: %w", err)
		}

		if len(datasets) == 0 {
			fmt.Printf("No datasets found in pool '%s'\n", poolName)
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tTYPE\tUSED\tAVAILABLE\tMOUNTPOINT\tCOMPRESSION")
		fmt.Fprintln(w, "----\t----\t----\t---------\t----------\t-----------")

		for _, ds := range datasets {
			compression := "-"
			if ds.Compression != nil {
				if val, ok := ds.Compression["value"].(string); ok {
					compression = val
				}
			}
			mountpoint := ds.Mountpoint
			if mountpoint == "" {
				mountpoint = "-"
			}

			// Extract used space
			used := "-"
			if ds.Used != nil {
				if val, ok := ds.Used["parsed"].(float64); ok {
					used = formatBytes(int64(val))
				}
			}

			// Extract available space
			available := "-"
			if ds.Available != nil {
				if val, ok := ds.Available["parsed"].(float64); ok {
					available = formatBytes(int64(val))
				}
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				ds.ID, ds.Type, used, available, mountpoint, compression)
		}
		w.Flush()

		return nil
	},
}

// formatBytes converts bytes to human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func init() {
	listCmd.Flags().StringVar(&server, "server", "", "TrueNAS server URL (e.g., https://192.168.1.100)")
	listCmd.Flags().StringVar(&token, "token", "", "TrueNAS API token")
}
