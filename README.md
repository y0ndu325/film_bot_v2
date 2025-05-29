# FilmBot

Telegram бот для управления списком фильмов.

## Функциональность

- Добавление фильмов в список
- Просмотр списка фильмов
- Случайный выбор фильма
- Удаление фильмов из списка

## Установка и запуск

### Локальная разработка

1. Клонируйте репозиторий:
```bash
git clone https://github.com/yourusername/filmBot.git
cd filmBot
```

2. Создайте файл `.env` на основе `.env.example`:
```bash
cp .env.example .env
```

3. Заполните переменные окружения в `.env`:
- `BOT_TOKEN` - токен вашего Telegram бота
- Настройки базы данных PostgreSQL

4. Установите зависимости:
```bash
go mod download
```

5. Запустите приложение:
```bash
go run main.go
```

### Деплой на Railway

1. Создайте новый проект на Railway
2. Подключите ваш GitHub репозиторий
3. Добавьте следующие переменные окружения:
   - `BOT_TOKEN`
   - `DB_HOST`
   - `DB_PORT`
   - `DB_USER`
   - `DB_PASSWORD`
   - `DB_NAME`
   - `DB_SSL_MODE`

## Структура проекта

```
.
├── cmd/            # Точка входа приложения
├── internal/       # Внутренний код приложения
│   ├── config/    # Конфигурация
│   ├── database/  # Работа с базой данных
│   ├── handlers/  # Обработчики Telegram
│   ├── models/    # Модели данных
│   └── service/   # Бизнес-логика
├── pkg/           # Публичные пакеты
└── assets/        # Статические файлы
```

## Технологии

- Go 1.21+
- PostgreSQL
- Telegram Bot API
- GORM 