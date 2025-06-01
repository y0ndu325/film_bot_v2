package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"filmBot/internal/config"
	"filmBot/internal/database"
	"filmBot/internal/handlers"
	"filmBot/internal/service"
)

func main() {
	// 1. Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	if cfg.BotToken == "" {
		log.Fatal("BOT_TOKEN environment variable is required")
	}

	// 2. Инициализируем базу данных
	db, err := database.NewDatabase(cfg.DBConfig.DSN())
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 3. Создаём экземпляр бота
	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}
	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	// 4. Очищаем очередь накопившихся апдейтов (если они были)
	//    Вместо bot.Request используем стандартный GetUpdates с Offset = -1.
	//    После этого Telegram вернёт все накопившиеся апдейты, и мы их просто проигнорируем.
	clearConfig := tgbotapi.NewUpdate(-1)
	clearConfig.Timeout = 0
	if _, err := bot.GetUpdates(clearConfig); err != nil {
		log.Printf("Error clearing updates queue: %v", err)
	}

	// 5. Инициализируем сервис и обработчик
	movieService := service.New(db)
	handler := handlers.New(bot, movieService, cfg)

	// 6. Запускаем HTTP-сервер для проверки здоровья (healthcheck)
	go func() {
		http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		})
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// 7. Настраиваем получение обновлений от Telegram
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60
	// убираем Offset = -1, потому что мы уже очистили очередь выше
	// ниже Telegram сам начинает слать новые апдейты начиная с последнего
	updateConfig.AllowedUpdates = []string{"message", "callback_query"}

	updates := bot.GetUpdatesChan(updateConfig)

	// 8. Готовимся к корректному завершению при SIGINT (Ctrl+C) или SIGTERM
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 9. Обработчик обновлений от Telegram
	go func() {
		for update := range updates {
			// Проверяем, что пришло именно текстовое сообщение или callback_query
			if update.Message != nil {
				if err := handler.HandleMessage(&update); err != nil {
					log.Printf("Error handling message: %v", err)
					// Отправляем пользователю сообщение об ошибке
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Произошла ошибка при обработке сообщения. Попробуйте еще раз.")
					if _, sendErr := bot.Send(msg); sendErr != nil {
						log.Printf("Error sending error message: %v", sendErr)
					}
				}
			} else if update.CallbackQuery != nil {
				// Если это callback_query (нажатие на inline-кнопку)
				if err := handler.HandleCallback(&update); err != nil {
					log.Printf("Error handling callback_query: %v", err)
					// Отправляем пользователю ошибку в чат того же сообщения
					chatID := update.CallbackQuery.Message.Chat.ID
					msg := tgbotapi.NewMessage(chatID, "Произошла ошибка при обработке нажатия. Попробуйте еще раз.")
					if _, sendErr := bot.Send(msg); sendErr != nil {
						log.Printf("Error sending callback error message: %v", sendErr)
					}
				}
				// Обязательно подтверждаем обработку callback (чтобы убрать "часики" в клиенте)
				ack := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
				if _, ackErr := bot.Request(ack); ackErr != nil {
					log.Printf("Error sending callback acknowledgement: %v", ackErr)
				}
			}
			// Если пришёл какой-то другой тип апдейта — просто пропускаем
		}
	}()

	// 10. Ожидаем сигнала завершения
	<-sigChan
	log.Println("Shutting down...")
}