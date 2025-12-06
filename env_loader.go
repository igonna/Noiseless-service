package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func loadEnvFile() {
	if _, err := os.Stat(".env"); err != nil {
		return
	}
	if err := godotenv.Load(); err != nil {
		log.Printf("warning: failed to load .env file: %v", err)
	}
}
