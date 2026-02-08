package cmd

import (
	"fmt"

	"lightweight-php/manager"

	"github.com/spf13/cobra"
)

var phpCmd = &cobra.Command{
	Use:   "php",
	Short: "Manage PHP installations",
	Long:  "Install and manage PHP versions from Remi repository",
}

var phpInstallCmd = &cobra.Command{
	Use:   "install [version]",
	Short: "Install a PHP version from Remi",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		version := args[0]
		pm, err := manager.NewPackageManager()
		if err != nil {
			fmt.Printf("Error initializing package manager: %v\n", err)
			return
		}
		if err := pm.InstallPHP(version); err != nil {
			fmt.Printf("Error installing PHP: %v\n", err)
			return
		}
		fmt.Printf("PHP %s installed successfully\n", version)
	},
}

var phpListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed PHP versions",
	Run: func(cmd *cobra.Command, args []string) {
		pm, err := manager.NewPackageManager()
		if err != nil {
			fmt.Printf("Error initializing package manager: %v\n", err)
			return
		}
		versions, err := pm.ListInstalledPHP()
		if err != nil {
			fmt.Printf("Error listing PHP versions: %v\n", err)
			return
		}
		for _, v := range versions {
			fmt.Printf("PHP %s\n", v)
		}
	},
}

func init() {
	phpCmd.AddCommand(phpInstallCmd)
	phpCmd.AddCommand(phpListCmd)
}
