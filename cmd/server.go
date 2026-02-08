package cmd

import (
	"fmt"
	"log"
	"net/http"

	"lightweight-php/api"

	"github.com/spf13/cobra"
)

var (
	serverHost string
	serverPort int
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the REST API server",
	Long:  "Start the REST API server for managing PHP-FPM pools and PHP installations",
	Run: func(cmd *cobra.Command, args []string) {
		router, err := api.NewRouter()
		if err != nil {
			log.Fatalf("Failed to initialize router: %v", err)
		}
		addr := fmt.Sprintf("%s:%d", serverHost, serverPort)
		log.Printf("Starting server on %s", addr)
		if err := http.ListenAndServe(addr, router); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	serverCmd.Flags().StringVarP(&serverHost, "host", "H", "0.0.0.0", "Server host")
	serverCmd.Flags().IntVarP(&serverPort, "port", "p", 8080, "Server port")
}
