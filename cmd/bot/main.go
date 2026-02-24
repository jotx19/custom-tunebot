package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

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

	b, err := bot.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	bot.StartHealthServer()

	if err := b.Start(); err != nil {
		log.Fatal(err)
	}
	log.Println("Bot running. Ctrl+C to stop.")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	_ = b.Close()
}
