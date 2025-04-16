package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var zub bool = false
var zub1 bool = false

// ..................................................................................................................... Сервер
func main() {
	go RunTelegramBot()

	r := gin.Default()

	// Настройка CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:3000"},
		AllowMethods: []string{"GET", "POST"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept"},
	}))

	// Маршруты API
	r.GET("/trackers", GetTrackers)
	r.POST("/trackers/update", UpdateTracker)
	r.POST("/trackers/month", GetTrackersByMonth)
	r.POST("/trackers/updatetelega", UpdateTrackerTelega)
	r.GET("/ws", handleWebSocket) // Новый маршрут для WebSocket

	// Запуск сервера
	r.Run(":8080")
}

// ..................................................................................................................... Запуск телеграмм бота
func RunTelegramBot() {
	// Получаем токен из переменной окружения
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("Ошибка: TELEGRAM_BOT_TOKEN не установлен. Укажи токен в переменных окружения.")
	}

	// Создаем экземпляр бота
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatalf("Ошибка при создании бота: %v", err)
	}

	bot.Debug = true
	log.Printf("Бот успешно авторизован как @%s", bot.Self.UserName)

	// Настраиваем обновления
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// Получаем канал обновлений
	updates := bot.GetUpdatesChan(u)

	// Создаем постоянную клавиатуру
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Хотьба"),
			tgbotapi.NewKeyboardButton("Карнитин"),
			tgbotapi.NewKeyboardButton("Кофе"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Туалет"),
			tgbotapi.NewKeyboardButton("Зубы"),
			tgbotapi.NewKeyboardButton("Витамины"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Мучное"),
			tgbotapi.NewKeyboardButton("Голова"),
			tgbotapi.NewKeyboardButton("Мазь"),
		),
	)

	// Обрабатываем входящие сообщения
	for update := range updates {
		if update.Message == nil { // Игнорируем обновления без сообщений
			continue
		}

		// Логируем сообщение
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		var msg tgbotapi.MessageConfig

		// Проверяем, начинается ли сообщение с "Работа:"
		if strings.HasPrefix(update.Message.Text, "Работа") {
			// Извлекаем текст после "Работа:"
			str := strings.TrimSpace(strings.TrimPrefix(update.Message.Text, "Работа"))
			if str == "" {
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Пожалуйста, укажи текст после 'Работа'")
			} else {
				if err := sendUpdate(9, str); err != nil {
					log.Printf("Ошибка при отправке POST-запроса: %v", err)
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при обновлении трекера.")
				} else {
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Работа записана: %s 👍", str))
				}
			}
		} else if strings.HasPrefix(update.Message.Text, "1Работа") {
			// Извлекаем текст после "Работа:"
			str := strings.TrimSpace(strings.TrimPrefix(update.Message.Text, "1Работа"))
			if str == "" {
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Пожалуйста, укажи текст после '1Работа'")
			} else {
				if err := sendUpdate(10, str); err != nil {
					log.Printf("Ошибка при отправке POST-запроса: %v", err)
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при обновлении трекера.")
				} else {
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Работа записана: %s 👍", str))
				}
			}
		} else {
			// Обрабатываем любое сообщение
			switch update.Message.Text {
			case "Хотьба":
				err1 := sendUpdate(1, "")
				if err1 != nil {
					log.Printf("Ошибка при отправке POST-запроса: %v", err)
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при обновлении трекера.")
					return
				}
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "ok!")
			case "Карнитин":
				err1 := sendUpdate(2, "")
				if err1 != nil {
					log.Printf("Ошибка при отправке POST-запроса: %v", err)
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при обновлении трекера.")
					return
				}
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "ok!")
			case "Кофе":
				err1 := sendUpdate(3, "")
				if err1 != nil {
					log.Printf("Ошибка при отправке POST-запроса: %v", err)
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при обновлении трекера.")
					return
				}
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "ok!")
			case "Туалет":
				err1 := sendUpdate(30, "")
				if err1 != nil {
					log.Printf("Ошибка при отправке POST-запроса: %v", err)
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при обновлении трекера.")
					return
				}
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "ok!")
			case "Зубы":
				err1 := sendUpdate(6, "")
				if err1 != nil {
					log.Printf("Ошибка при отправке POST-запроса: %v", err)
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при обновлении трекера.")
					return
				}
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "ok!")
			case "Витамины":
				err1 := sendUpdate(24, "")
				if err1 != nil {
					log.Printf("Ошибка при отправке POST-запроса: %v", err)
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при обновлении трекера.")
					return
				}
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "ok!")
			case "Мучное":
				err1 := sendUpdate(27, "")
				if err1 != nil {
					log.Printf("Ошибка при отправке POST-запроса: %v", err)
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при обновлении трекера.")
					return
				}
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "ok!")
			case "Голова":
				err1 := sendUpdate(23, "")
				if err1 != nil {
					log.Printf("Ошибка при отправке POST-запроса: %v", err)
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при обновлении трекера.")
					return
				}
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "ok!")
			case "Мазь":
				err1 := sendUpdate(29, "")
				if err1 != nil {
					log.Printf("Ошибка при отправке POST-запроса: %v", err)
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при обновлении трекера.")
					return
				}
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "ok!")
			default:
				// При любом первом сообщении показываем клавиатуру
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Привет! Я готов к работе. Выбери действие:")
			}
		}

		// Устанавливаем клавиатуру
		msg.ReplyMarkup = keyboard

		// Отправляем сообщение
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Ошибка при отправке сообщения: %v", err)
		}
	}
}

// ..................................................................................................................... Отправляет POST-запрос на сервер
func sendUpdate(id int, str string) error {
	// Структура запроса
	type Request struct {
		Day   int    `json:"day"`   // День года (1–365)
		Index int    `json:"index"` // Индекс кнопки
		State int    `json:"state"` // Новое состояние
		Text  string `json:"text"`  // Текст (если есть)
		Month int    `json:"month"` // Месяц (0–11), пока не используется
	}

	// Получаем текущий день
	day := time.Now().Day()
	month := int(time.Now().Month()) - 1

	if id == 3 && zub == false && zub1 == false {
		zub = true
		goto xxx
	}
	if id == 3 && zub == true && zub1 == false {
		id = 4
		zub1 = true
		goto xxx
	}
	if id == 3 && zub == true && zub1 == true {
		id = 5
		goto xxx
	}

xxx:
	// Формируем данные для запроса
	reqData := Request{
		Day:   day, // Сегодняшний день (число)
		Index: id,  // Фиксированное значение
		State: 1,   // Ты не указал значение для State, поставил 0 как заглушку
		Text:  str, // Пустая строка
		Month: month,
	}

	// Преобразуем в JSON
	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return err
	}

	// Создаем POST-запрос
	resp, err := http.Post("http://localhost:8080/trackers/updatetelega", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("сервер вернул ошибку: %s", resp.Status)
	}

	return nil
}
