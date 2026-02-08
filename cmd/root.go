package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "lightweight-php",
	Short: "PHP-FPM pool manager with REST API",
	Long:  "A CLI tool to manage PHP-FPM pools per user and install PHP versions from Remi repository",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(poolCmd)
	rootCmd.AddCommand(phpCmd)
}
