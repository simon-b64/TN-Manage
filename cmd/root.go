package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "tnmanage",
	Short: "Manage datasets and NFS shares on TrueNAS",
	Long:  `A command line tool to manage datasets and NFS shares on TrueNAS systems.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(clearCmd)
	rootCmd.AddCommand(configCmd)
}
