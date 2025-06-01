package handlers

import (
	"filmBot/internal/config"
	"filmBot/internal/service"
	"fmt"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler struct {
	bot     *tgbotapi.BotAPI
	service *service.MovieService
	config  *config.Config
	states  map[int64]string
}

func (h *Handler) HandleCallback(update *tgbotapi.Update) error {
	if update.CallbackQuery == nil {
		return nil
	}

	chatID := update.CallbackQuery.Message.Chat.ID
	data := update.CallbackQuery.Data

	switch data {
	case "delete":
		return h.handleDelete(chatID)
	case "list":
		return h.handleList(chatID)
	case "random":
		return h.handleRandom(chatID)
	default:
		msg := tgbotapi.NewMessage(chatID, "Неизвестная команда")
		_, err := h.bot.Send(msg)
		return err
	}
}

func New(bot *tgbotapi.BotAPI, service *service.MovieService, config *config.Config) *Handler {
	return &Handler{
		bot:     bot,
		service: service,
		config:  config,
		states:  make(map[int64]string),
	}
}

func (h *Handler) HandleMessage(update *tgbotapi.Update) error {
	if update.Message == nil {
		return nil
	}

	chatID := update.Message.Chat.ID
	text := update.Message.Text

	switch {
	case text == "/start":
		return h.handleStart(chatID)
	case text == "Список" || text == "список":
		return h.handleList(chatID)
	case text == "Рандом" || text == "рандом":
		return h.handleRandom(chatID)
	case text == "Удалить" || text == "удалить":
		return h.handleDelete(chatID)
	default:
		return h.handleDefault(chatID, text)
	}
}

func (h *Handler) handleStart(chatID int64) error {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Список"),
			tgbotapi.NewKeyboardButton("Рандом"),
			tgbotapi.NewKeyboardButton("Удалить"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, "Привет! Отправь мне название фильма, и я сохраню его.")
	msg.ReplyMarkup = keyboard
	_, err := h.bot.Send(msg)
	return err
}

func (h *Handler) handleList(chatID int64) error {
	movies, err := h.service.GetMovies()
	if err != nil {
		return err
	}

	if len(movies) == 0 {
		msg := tgbotapi.NewMessage(chatID, "Список фильмов пуст")
		_, err := h.bot.Send(msg)
		return err
	}

	var response string
	for _, movie := range movies {
		response += fmt.Sprintf("-%s\n", movie.Title)
	}

	// Разбиваем сообщение на части, если оно слишком длинное
	const maxLength = 4000
	if len(response) > maxLength {
		response = response[:maxLength] + "...\n(список обрезан из-за ограничений Telegram)"
	}

	msg := tgbotapi.NewMessage(chatID, "Список фильмов:\n"+response)
	_, err = h.bot.Send(msg)
	return err
}

func (h *Handler) handleRandom(chatID int64) error {
	movie, err := h.service.GetRandomMovie()
	if err != nil {
		if err == service.ErrNoMovies {
			msg := tgbotapi.NewMessage(chatID, "Список фильмов пуст.")
			_, err := h.bot.Send(msg)
			return err
		}
		return err
	}

	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Будем смотреть это!: %s", movie.Title))
	_, err = h.bot.Send(msg)
	return err
}

func (h *Handler) handleDelete(chatID int64) error {
	movies, err := h.service.GetMovies()
	if err != nil {
		return err
	}
	if len(movies) == 0 {
		msg := tgbotapi.NewMessage(chatID, "Список фильмов пуст.")
		_, err := h.bot.Send(msg)
		return err
	}

	var response string
	for i, movie := range movies {
		response += fmt.Sprintf("%d. %s\n", i+1, movie.Title)
	}

	msg := tgbotapi.NewMessage(chatID, "Выберите фильм для удаления:\n"+response)
	_, err = h.bot.Send(msg)
	if err != nil {
		return err
	}

	h.states[chatID] = "wait_del"
	return nil
}

func (h *Handler) handleDeleteConfirmation(chatID int64, text string) error {
	number, err := strconv.Atoi(text)
	if err != nil || number <= 0 {
		msg := tgbotapi.NewMessage(chatID, "Введите корректный номер фильма (положительное число).")
		_, err := h.bot.Send(msg)
		return err
	}

	movie, err := h.service.DeleteMovie(number)
	if err != nil {
		if err == service.ErrInvalidIndex {
			msg := tgbotapi.NewMessage(chatID, "Некорректный номер фильма.")
			_, err := h.bot.Send(msg)
			return err
		}
		return err
	}

	photo := tgbotapi.NewPhoto(chatID, tgbotapi.FilePath("assets/del_image.jpg"))
	photo.Caption = fmt.Sprintf("Фильм '%s' удален.", movie.Title)
	_, err = h.bot.Send(photo)
	if err != nil {
		return err
	}

	delete(h.states, chatID)
	return nil
}

func (h *Handler) handleDefault(chatID int64, text string) error {
	if state, exists := h.states[chatID]; exists && state == "wait_del" {
		return h.handleDeleteConfirmation(chatID, text)
	}

	if err := h.service.AddMovie(text); err != nil {
		return err
	}

	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Фильм '%s' сохранен.", text))
	_, err := h.bot.Send(msg)
	return err
}
