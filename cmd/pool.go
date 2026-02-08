package cmd

import (
	"fmt"

	"lightweight-php/manager"

	"github.com/spf13/cobra"
)

var poolCmd = &cobra.Command{
	Use:   "pool",
	Short: "Manage PHP-FPM pools",
	Long:  "Create, delete, and manage PHP-FPM pools for users",
}

var poolCreateCmd = &cobra.Command{
	Use:   "create [username]",
	Short: "Create a PHP-FPM pool for a user",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		username := args[0]
		phpVersion, _ := cmd.Flags().GetString("php-version")
		provider, _ := cmd.Flags().GetString("provider")
		
		if provider == "" {
			provider = "remi"
		}
		
		pm, err := manager.NewPoolManager()
		if err != nil {
			fmt.Printf("Error initializing pool manager: %v\n", err)
			return
		}
		if err := pm.CreatePool(username, phpVersion, provider); err != nil {
			fmt.Printf("Error creating pool: %v\n", err)
			return
		}
		fmt.Printf("Pool created for user: %s with PHP %s (provider: %s)\n", username, phpVersion, provider)
	},
}

var poolDeleteCmd = &cobra.Command{
	Use:   "delete [username]",
	Short: "Delete a PHP-FPM pool for a user",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		username := args[0]
		pm, err := manager.NewPoolManager()
		if err != nil {
			fmt.Printf("Error initializing pool manager: %v\n", err)
			return
		}
		if err := pm.DeletePool(username); err != nil {
			fmt.Printf("Error deleting pool: %v\n", err)
			return
		}
		fmt.Printf("Pool deleted for user: %s\n", username)
	},
}

var poolListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all PHP-FPM pools",
	Run: func(cmd *cobra.Command, args []string) {
		pm, err := manager.NewPoolManager()
		if err != nil {
			fmt.Printf("Error initializing pool manager: %v\n", err)
			return
		}
		pools, err := pm.ListPools()
		if err != nil {
			fmt.Printf("Error listing pools: %v\n", err)
			return
		}
		for _, pool := range pools {
			fmt.Printf("User: %s, PHP Version: %s, Provider: %s, Status: %s\n", pool.User, pool.PHPVersion, pool.Provider, pool.Status)
		}
	},
}

func init() {
	poolCmd.AddCommand(poolCreateCmd)
	poolCmd.AddCommand(poolDeleteCmd)
	poolCmd.AddCommand(poolListCmd)
	poolCreateCmd.Flags().String("php-version", "8.2", "PHP version to use")
	poolCreateCmd.Flags().String("provider", "remi", "PHP provider (remi, lsphp, alt-php, docker)")
}
