package main

import (
	"filmBot/internal/config"
	"filmBot/internal/database"
	"filmBot/internal/handlers"
	"filmBot/internal/service"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if cfg.BotToken == "" {
		log.Fatal("BOT_TOKEN environment variable is required")
	}

	// Initialize database
	db, err := database.NewDatabase(cfg.DBConfig.DSN())
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize bot
	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}
	bot.Debug = true

	// Initialize service
	movieService := service.New(db)

	// Initialize handler
	handler := handlers.New(bot, movieService, cfg)

	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Start HTTP server for health check
	go func() {
		http.HandleFunc("/kaithheathcheck", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for update := range updates {
			if err := handler.HandleMessage(&update); err != nil {
				log.Printf("Error handling message: %v", err)
			}
		}
	}()

	<-sigChan
	log.Println("Shutting down...")
}
