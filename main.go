package main

import (
	"log"
	"os"

	"lightweight-php/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
