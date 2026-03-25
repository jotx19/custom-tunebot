package main

import (
	"log"
	"musicbot/internal/bot"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env not loaded, using system env")
	}
	cfg, err := bot.LoadConfigFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	if err := bot.RegisterCommands(cfg); err != nil {
		log.Fatal("Failed to register commands:", err)
	}
	log.Println("Commands registered successfully!")
}
