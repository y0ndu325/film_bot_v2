package main

import (
	"filmBot/internal/config"
	"filmBot/internal/database"
	"filmBot/internal/handlers"
	"filmBot/internal/service"
	"log"

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

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if err := handler.HandleMessage(&update); err != nil {
			log.Printf("Error handling message: %v", err)
		}
	}
}
