package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Movie struct {
	ID    uint   `gorm:"primaryKey"`
	Title string `gorm:"uniqueIndex"`
}

var forbiddenWords = []string {
	"очень страшное кино",
	"очень страшноекино",
	"оченьстрашное кино",
	"оченьстрашноекино",
	"очень страшное кино 1",
	"очень страшноекино1",
	"оченьстрашное кино 1",
	"оченьстрашноекино 1",
	"очень страшное кино 2",
	"очень страшноекино2",
	"оченьстрашное кино 2",
	"оченьстрашноекино 2",
	"очень страшное кино 3",
	"очень страшноекино3",
	"оченьстрашное кино 3",
	"оченьстрашноекино 3",
	"очень страшное кино 4",
	"очень страшноекино4",
	"оченьстрашное кино 4",
	"оченьстрашноекино 4",
}

func containsForbiddenWord(title string) bool {
	title = strings.ToLower(title)
	for _, word := range forbiddenWords {
		if strings.Contains(title, strings.ToLower(word)) {
			return true
		}
	}
	return false
}

var userStates = make(map[int64]string)

func main() {
	db, err := gorm.Open(sqlite.Open("movies.db"),&gorm.Config{})
	if err != nil {
		log.Fatalf("нет подключения к базе данных: %v", err)
	}
	db.AutoMigrate(&Movie{})

	
	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
    	log.Fatalf("Токен бота не найден в переменных окружения")
}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatalf("не удалось создать бота: %v", err)
	}
	bot.Debug = true

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Список"),
			tgbotapi.NewKeyboardButton("Рандом"),
			tgbotapi.NewKeyboardButton("Удалить"),
		),
	)

	log.Printf("авторизация бота: %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		chatID := update.Message.Chat.ID
		text := update.Message.Text

		switch {
			case text == "/start":
				msg := tgbotapi.NewMessage(chatID, "Привет! Отправь мне название фильма, и я сохраню его.")
				msg.ReplyMarkup = keyboard
				bot.Send(msg)

			case text == "Список":
				var movies []Movie
				db.Find(&movies)
				if len(movies) == 0 {
					msg := tgbotapi.NewMessage(chatID, "Список фильмов пуст.")
					bot.Send(msg)
					continue
				}
				var response strings.Builder
				response.WriteString("Список фильмов:\n")
				for _, movie := range movies {
					response.WriteString(fmt.Sprintf("- %s\n", movie.Title))
				}
				msg := tgbotapi.NewMessage(chatID, response.String())
				bot.Send(msg)

			case text == "/clear":
				db.Where("1 = 1").Delete(&Movie{})
				msg := tgbotapi.NewMessage(chatID, "Список фильмов очищен.")
				bot.Send(msg)

			case text == "Рандом":
				var movies []Movie
				db.Find(&movies)
				if len(movies) == 0 {
					msg := tgbotapi.NewMessage(chatID, "Список фильмов пуст.")
					bot.Send(msg)
					continue
				}

				rand.Seed(time.Now().UnixNano())
				randomIndex := rand.Intn(len(movies))
				randomMovie := movies[randomIndex]
//delete random movie
				db.Delete(&randomMovie)
				msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("будем смотреть это!: %s", randomMovie.Title))
				bot.Send(msg)

			case text == "Удалить":
				var movies []Movie
				db.Find(&movies)
				if len(movies) == 0 {
					msg := tgbotapi.NewMessage(chatID, "Список фильмов пуст.")
					bot.Send(msg)
					break
				}

				var response strings.Builder
				response.WriteString("Выберите фильм для удаления:\n")
				for i, movie := range movies {
					response.WriteString(fmt.Sprintf("%d. %s\n", i+1, movie.Title))
				}
				msg := tgbotapi.NewMessage(chatID, response.String())
				bot.Send(msg)

				userStates[chatID] = "wait_del"

			default:
				if state, exists := userStates[chatID]; exists && state == "wait_del" {
					number, err := strconv.Atoi(text)
					if err != nil {
						msg := tgbotapi.NewMessage(chatID, "Введите корректный номер фильма.")
						bot.Send(msg)
						break
					}
					var movies []Movie
					db.Find(&movies)
					
					if number < 1 || number > len(movies) {
						msg := tgbotapi.NewMessage(chatID, "Некорректный номер фильма.")
						bot.Send(msg)
					}else {
						movieToDelete := movies[number-1]
						db.Delete(&movieToDelete)

						photo := tgbotapi.NewPhoto(chatID, tgbotapi.FilePath("assets/del_image.jpg"))
						photo.Caption = fmt.Sprintf("Фильм '%s' удален.", movieToDelete.Title)
						bot.Send(photo)
					}
					
					delete(userStates, chatID)
				}else {
					if containsForbiddenWord(text) {
						msg := tgbotapi.NewMessage(chatID, "увы Ваня....... мы не будем это смотреть.")
						bot.Send(msg)
						continue
					}
	
					movie := Movie{Title: text}
					result := db.Create(&movie)
					if result.Error != nil {
						msg := tgbotapi.NewMessage(chatID, "Произошла ошибка при сохранении фильма.")
						bot.Send(msg)
					}else {
						msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Фильм '%s' сохранен.", text))
						bot.Send(msg)
					}
				}
				

			}
			

	}
}
