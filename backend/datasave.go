package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Button struct {
	State int    `json:"state"`
	Text  string `json:"text,omitempty"`
}

var db *sql.DB

// WebSocket-клиенты и настройки
var (
	wsClients  = make(map[*websocket.Conn]bool)
	wsUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true }, // Для тестов
	}
)

// .................................................................................................................... Подключение к БД
func init() {
	//................................................................................................................. для Компа локально подключение
	var err error
	db, err = sql.Open("postgres", "user=postgres password=12345 host=localhost port=5432 dbname=Ksandr_Test_DB_01 sslmode=disable")
	if err != nil {
		fmt.Println("Ошибка подключения к базе данных: ", err)
	}
	fmt.Println("----------------------- [MY-LOG] → Успешное подключение к базе данных")

	//................................................................................................................. для Docker подключение
	//var err error
	//db, err = sql.Open("postgres", "user=postgres password=12345 host=db port=5432 dbname=Ksandr_Test_DB_01 sslmode=disable")
	//if err != nil {
	//	fmt.Println("Ошибка подключения к базе данных: ", err)
	//	return // Или обработайте ошибку, например, os.Exit(1)
	//}
	//fmt.Println("----------------------- [MY-LOG] → Успешное подключение к базе данных")
}

// .................................................................................................................... Загрузка актуального месяца в начале
func GetTrackers(c *gin.Context) {
	// Получение текущего года и месяца
	now := time.Now()
	currentYear := now.Year()
	currentMonth := int(now.Month())

	// Запрос строк за текущий месяц из новой таблицы
	rows, err := db.Query(`
		SELECT * FROM tbl_calendar_2025_prod_v2
		WHERE EXTRACT(YEAR FROM data_state) = $1 
		AND EXTRACT(MONTH FROM data_state) = $2 
		ORDER BY day
	`, currentYear, currentMonth)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	data := make(map[int][]Button)
	for rows.Next() {
		var day int
		var dataState time.Time // Для столбца data_state
		var btns [31]Button     // Увеличиваем до 34, так как в новой таблице больше столбцов

		// Считываем все столбцы из новой таблицы
		err = rows.Scan(&day, &dataState,
			&btns[0].State, &btns[0].Text, // ves_state, ves_text
			&btns[1].State,                // hotba_state
			&btns[2].State,                // lkarnitin_state
			&btns[3].State,                // kofe01_state
			&btns[4].State,                // kofe02_state
			&btns[5].State,                // kofe03_state
			&btns[6].State,                // zub01_state
			&btns[7].State,                // zub02_state
			&btns[8].State,                // rabotastatus_state (новый столбец)
			&btns[9].State, &btns[9].Text, // rabota_state, rabota_text
			&btns[10].State, &btns[10].Text, // rabotautro_state, rabotautro_text
			&btns[11].State, &btns[11].Text, // rabotaden_state, rabotaden_text
			&btns[12].State, &btns[12].Text, // rabotavecher_state, rabotavecher_text
			&btns[13].State, &btns[13].Text, // rabotasum_state, rabotasum_text
			&btns[14].State,                 // xxxs_state
			&btns[15].State, &btns[15].Text, // nachalosna_state, nachalosna_text
			&btns[16].State, &btns[16].Text, // dlinasna_state, dlinasna_text
			&btns[17].State, // krasotasna_state
			&btns[18].State, // xxxm_state
			&btns[19].State, // nastroenieutro_state
			&btns[20].State, // nastroenieden_state
			&btns[21].State, // alkogol_state
			&btns[22].State, // dush_state
			&btns[23].State, // golova_state
			&btns[24].State, // vitamin_state
			&btns[25].State, // fastfood_state
			&btns[26].State, // e1800_state
			&btns[27].State, // muchnoe_state
			&btns[28].State, // klava_state
			&btns[29].State, // maz_state
			&btns[30].State, // tualet_state
		)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		data[day] = btns[:]
	}

	c.JSON(200, data)
}

// .................................................................................................................... Запись в БД с сайта
func UpdateTracker(c *gin.Context) {
	fmt.Println("----------------------- [MY-LOG] → 01")
	var req struct {
		Day   int    `json:"day"`   // День года (1–365)
		Index int    `json:"index"` // Индекс кнопки
		State int    `json:"state"` // Новое состояние
		Text  string `json:"text"`  // Текст (если есть)
		Month int    `json:"month"` // Месяц (0–11)
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	fmt.Println("----------------------- [MY-LOG] → 02")
	// Проверка допустимого дня года (1–365, 2025 не високосный)
	if req.Day < 1 || req.Day > 365 {
		c.JSON(400, gin.H{"error": "Invalid day of year"})
		return
	}

	// Маппинг индексов на имена столбцов
	names := []string{
		"ves", "hotba", "lkarnitin", "kofe01", "kofe02", "kofe03",
		"zub01", "zub02", "rabotastatus", "rabota",
		"rabotautro", "rabotaden", "rabotavecher", "rabotasum", "xxxs",
		"nachalosna", "dlinasna", "krasotasna", "xxxm", "nastroenieutro",
		"nastroenieden", "alkogol", "dush", "golova", "vitamin",
		"fastfood", "e1800", "muchnoe", "klava", "maz", "tualet",
	}

	// Проверка допустимого индекса
	if req.Index < 0 || req.Index >= len(names) {
		c.JSON(400, gin.H{"error": "Invalid index"})
		return
	}

	// Список индексов, для которых нужно обновлять text
	updateTextIndices := map[int]bool{
		0:  true, // ves
		9:  true, // rabota
		10: true, // rabotautro
		11: true, // rabotaden
		12: true, // rabotavecher
		13: true, // rabotasum
		15: true, // nachalosna
		16: true, // dlinasna
	}

	// Формируем запрос в зависимости от индекса
	name := names[req.Index]
	var query string
	var args []interface{}
	if updateTextIndices[req.Index] {
		// Обновляем state и text
		query = fmt.Sprintf("UPDATE tbl_calendar_2025_prod_v2 SET %s_state = $1, %s_text = $2 WHERE day = $3", name, name)
		args = []interface{}{req.State, req.Text, req.Day}
	} else {
		// Обновляем только state
		query = fmt.Sprintf("UPDATE tbl_calendar_2025_prod_v2 SET %s_state = $1 WHERE day = $2", name)
		args = []interface{}{req.State, req.Day}
	}

	// Выполняем запрос
	_, err := db.Exec(query, args...)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Дополнительная логика для rabotasum (индекс 13)
	if req.Index == 13 {
		// Извлекаем данные из rabotautro_text, rabotaden_text, rabotavecher_text
		var rabotautroText, rabotadenText, rabotavecherText sql.NullString
		query = `
            SELECT rabotautro_text, rabotaden_text, rabotavecher_text 
            FROM tbl_calendar_2025_prod_v2 
            WHERE day = $1
        `
		err := db.QueryRow(query, req.Day).Scan(&rabotautroText, &rabotadenText, &rabotavecherText)
		if err != nil {
			if err == sql.ErrNoRows {
				// Если строки нет, можно либо создать её, либо пропустить
			} else {
				c.JSON(500, gin.H{"error": fmt.Sprintf("Ошибка при получении данных: %v", err)})
				return
			}
		}

		// Преобразуем строки в float64 и суммируем
		sum := 0.0
		for i, text := range []sql.NullString{rabotautroText, rabotadenText, rabotavecherText} {
			if text.Valid {
				// Удаляем пробелы в начале и конце строки
				cleanedText := strings.TrimSpace(text.String)
				// Проверяем, является ли строка пустой после очистки
				if cleanedText == "" {
					fmt.Printf("----------------------- [MY-LOG] → Текст пустой (после очистки) для индекса %d, добавляем 0.0\n", i+10)
					sum += 0.0
					continue
				}
				// Преобразуем очищенный текст в float64
				value, err := strconv.ParseFloat(cleanedText, 64)
				if err != nil {
					fmt.Printf("----------------------- [MY-LOG] → Ошибка преобразования текста в float64 для индекса %d: %v\n", i+10, err)
					continue // Пропускаем некорректное значение
				}
				fmt.Printf("----------------------- [MY-LOG] → float64 равен %.1f\n", value)
				sum += value
			} else {
				fmt.Printf("----------------------- [MY-LOG] → Текст NULL для индекса %d, добавляем 0.0\n", i+10)
				sum += 0.0
			}
		}

		// Преобразуем сумму обратно в строку
		sumStr := fmt.Sprintf("%.1f", sum)

		// Обновляем rabotasum_text
		query = "UPDATE tbl_calendar_2025_prod_v2 SET rabotasum_text = $1 WHERE day = $2"
		_, err = db.Exec(query, sumStr, req.Day)
		if err != nil {
			c.JSON(500, gin.H{"error": fmt.Sprintf("Ошибка при обновлении rabotasum_text: %v", err)})
			return
		}
	}

	// После обновления отправляем данные на фронтенд через WebSocket
	data, err := getTrackersData(db, req.Month)
	if err != nil {
		log.Printf("Ошибка получения данных для WebSocket: %v", err)
	} else {
		broadcastTrackersUpdate(data)
	}

	c.JSON(200, gin.H{"status": "updated"})
}

// .................................................................................................................... Переключение Месяца
func GetTrackersByMonth(c *gin.Context) {
	var req struct {
		Month int `json:"month"` // Месяц (0–11)
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Проверка допустимого месяца
	if req.Month < 0 || req.Month > 11 {
		c.JSON(400, gin.H{"error": "Invalid month"})
		return
	}

	// Запрос данных за указанный месяц
	rows, err := db.Query(`
		SELECT * FROM tbl_calendar_2025_prod_v2 
		WHERE EXTRACT(MONTH FROM data_state) = $1 
		ORDER BY day
	`, req.Month+1) // PostgreSQL считает месяцы с 1 (январь = 1), а фронтенд отправляет 0–11
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	data := make(map[int][]Button)
	for rows.Next() {
		var day int
		var dataState time.Time
		var btns [31]Button // 31 кнопка, как в DayRow

		err = rows.Scan(&day, &dataState,
			&btns[0].State, &btns[0].Text, // ves_state, ves_text
			&btns[1].State,                // hotba_state
			&btns[2].State,                // lkarnitin_state
			&btns[3].State,                // kofe01_state
			&btns[4].State,                // kofe02_state
			&btns[5].State,                // kofe03_state
			&btns[6].State,                // zub01_state
			&btns[7].State,                // zub02_state
			&btns[8].State,                // rabotastatus_state
			&btns[9].State, &btns[9].Text, // rabota_state, rabota_text
			&btns[10].State, &btns[10].Text, // rabotautro_state, rabotautro_text
			&btns[11].State, &btns[11].Text, // rabotaden_state, rabotaden_text
			&btns[12].State, &btns[12].Text, // rabotavecher_state, rabotavecher_text
			&btns[13].State, &btns[13].Text, // rabotasum_state, rabotasum_text
			&btns[14].State,                 // xxxs_state
			&btns[15].State, &btns[15].Text, // nachalosna_state, nachalosna_text
			&btns[16].State, &btns[16].Text, // dlinasna_state, dlinasna_text
			&btns[17].State, // krasotasna_state
			&btns[18].State, // xxxm_state
			&btns[19].State, // nastroenieutro_state
			&btns[20].State, // nastroenieden_state
			&btns[21].State, // alkogol_state
			&btns[22].State, // dush_state
			&btns[23].State, // golova_state
			&btns[24].State, // vitamin_state
			&btns[25].State, // fastfood_state
			&btns[26].State, // e1800_state
			&btns[27].State, // muchnoe_state
			&btns[28].State, // klava_state
			&btns[29].State, // maz_state
			&btns[30].State, // tualet_state
		)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		data[day] = btns[:]
	}

	c.JSON(200, data)
}

// .................................................................................................................... Запись в БД с телеги
func UpdateTrackerTelega(c *gin.Context) {
	var req struct {
		Day   int    `json:"day"`   // День месяца (1–31)
		Index int    `json:"index"` // Индекс кнопки
		State int    `json:"state"` // Новое состояние
		Text  string `json:"text"`  // Текст (если есть)
		Month int    `json:"month"` // Месяц (0–11)
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Преобразуем день месяца и месяц в день года
	dayOfYear, err := dayOfYear(req.Month, req.Day)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Проверка допустимого дня года (1–365, 2025 не високосный)
	if dayOfYear < 1 || dayOfYear > 365 {
		c.JSON(400, gin.H{"error": "Invalid day of year"})
		return
	}

	// Маппинг индексов на имена столбцов
	names := []string{
		"ves", "hotba", "lkarnitin", "kofe01", "kofe02", "kofe03",
		"zub01", "zub02", "rabotastatus", "rabota",
		"rabotautro", "rabotaden", "rabotavecher", "rabotasum", "xxxs",
		"nachalosna", "dlinasna", "krasotasna", "xxxm", "nastroenieutro",
		"nastroenieden", "alkogol", "dush", "golova", "vitamin",
		"fastfood", "e1800", "muchnoe", "klava", "maz", "tualet",
	}

	if req.Index < 0 || req.Index >= len(names) {
		c.JSON(400, gin.H{"error": "Invalid index"})
		return
	}

	updateTextIndices := map[int]bool{
		0: true, 9: true, 10: true, 11: true, 12: true, 13: true, 15: true, 16: true,
	}

	name := names[req.Index]
	var query string
	var args []interface{}
	if updateTextIndices[req.Index] {
		query = fmt.Sprintf("UPDATE tbl_calendar_2025_prod_v2 SET %s_state = $1, %s_text = $2 WHERE day = $3", name, name)
		args = []interface{}{req.State, req.Text, dayOfYear}
	} else {
		query = fmt.Sprintf("UPDATE tbl_calendar_2025_prod_v2 SET %s_state = $1 WHERE day = $2", name)
		args = []interface{}{req.State, dayOfYear}
	}

	_, err = db.Exec(query, args...)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// После обновления отправляем данные на фронтенд через WebSocket
	data, err := getTrackersData(db, req.Month)
	if err != nil {
		log.Printf("Ошибка получения данных для WebSocket: %v", err)
	} else {
		broadcastTrackersUpdate(data)
	}

	c.JSON(200, gin.H{"status": "updated"})
}

// .................................................................................................................... Вспомогательная функция для получения данных из БД
func getTrackersData(db *sql.DB, month int) (map[int][]Button, error) {
	// Определяем диапазон дней года для месяца
	startDay, endDay, err := monthDayRange(month)
	if err != nil {
		return nil, fmt.Errorf("ошибка вычисления диапазона дней: %v", err)
	}

	// Запрос с фильтрацией по диапазону дней
	query := "SELECT * FROM tbl_calendar_2025_prod_v2 WHERE day BETWEEN $1 AND $2 ORDER BY day"
	rows, err := db.Query(query, startDay, endDay)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса к базе: %v", err)
	}
	defer rows.Close()

	data := make(map[int][]Button)
	for rows.Next() {
		var day int
		var dataState time.Time
		var btns [31]Button // 31 кнопка, как в DayRow

		err = rows.Scan(&day, &dataState,
			&btns[0].State, &btns[0].Text, // ves_state, ves_text
			&btns[1].State,                // hotba_state
			&btns[2].State,                // lkarnitin_state
			&btns[3].State,                // kofe01_state
			&btns[4].State,                // kofe02_state
			&btns[5].State,                // kofe03_state
			&btns[6].State,                // zub01_state
			&btns[7].State,                // zub02_state
			&btns[8].State,                // rabotastatus_state
			&btns[9].State, &btns[9].Text, // rabota_state, rabota_text
			&btns[10].State, &btns[10].Text, // rabotautro_state, rabotautro_text
			&btns[11].State, &btns[11].Text, // rabotaden_state, rabotaden_text
			&btns[12].State, &btns[12].Text, // rabotavecher_state, rabotavecher_text
			&btns[13].State, &btns[13].Text, // rabotasum_state, rabotasum_text
			&btns[14].State,                 // xxxs_state
			&btns[15].State, &btns[15].Text, // nachalosna_state, nachalosna_text
			&btns[16].State, &btns[16].Text, // dlinasna_state, dlinasna_text
			&btns[17].State, // krasotasna_state
			&btns[18].State, // xxxm_state
			&btns[19].State, // nastroenieutro_state
			&btns[20].State, // nastroenieden_state
			&btns[21].State, // alkogol_state
			&btns[22].State, // dush_state
			&btns[23].State, // golova_state
			&btns[24].State, // vitamin_state
			&btns[25].State, // fastfood_state
			&btns[26].State, // e1800_state
			&btns[27].State, // muchnoe_state
			&btns[28].State, // klava_state
			&btns[29].State, // maz_state
			&btns[30].State, // tualet_state
		)
		if err != nil {
			return nil, err
		}
		data[day] = btns[:]
	}
	return data, nil
}

// .................................................................................................................... Обработка WebSocket-соединений
func handleWebSocket(c *gin.Context) {
	conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Ошибка обновления до WebSocket: %v", err)
		return
	}
	defer conn.Close()

	wsClients[conn] = true
	log.Printf("Клиент WebSocket подключен. Всего клиентов: %d", len(wsClients))

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			log.Printf("Клиент отключился: %v", err)
			delete(wsClients, conn)
			break
		}
	}
}

// .................................................................................................................... Отправка данных через WebSocket всем клиентам
func broadcastTrackersUpdate(data map[int][]Button) {
	payload, err := json.Marshal(data)
	if err != nil {
		log.Printf("Ошибка сериализации данных для WebSocket: %v", err)
		return
	}

	for conn := range wsClients {
		if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			log.Printf("Ошибка отправки через WebSocket: %v", err)
			conn.Close()
			delete(wsClients, conn)
		}
	}
	log.Printf("Данные отправлены %d клиентам через WebSocket", len(wsClients))
}

// .................................................................................................................... Вычисляет день года (1–365) для 2025 года на основе месяца (0–11) и дня месяца (1–31)
func dayOfYear(month, day int) (int, error) {
	// Массив с количеством дней в каждом месяце (2025 год не високосный)
	daysInMonth := []int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}

	// Проверка корректности месяца
	if month < 0 || month > 11 {
		return 0, fmt.Errorf("invalid month: %d (must be 0–11)", month)
	}

	// Проверка корректности дня
	if day < 1 || day > daysInMonth[month] {
		return 0, fmt.Errorf("invalid day: %d for month %d", day, month)
	}

	// Считаем день года
	dayOfYear := 0
	for i := 0; i < month; i++ {
		dayOfYear += daysInMonth[i]
	}
	dayOfYear += day

	return dayOfYear, nil
}

func monthDayRange(month int) (int, int, error) {
	// Массив с количеством дней в каждом месяце (2025 год не високосный)
	daysInMonth := []int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}

	// Проверка корректности месяца
	if month < 0 || month > 11 {
		return 0, 0, fmt.Errorf("invalid month: %d (must be 0–11)", month)
	}

	// Вычисляем начальный день года для месяца
	startDay, err := dayOfYear(month, 1)
	if err != nil {
		return 0, 0, err
	}

	// Вычисляем конечный день года для месяца
	endDay, err := dayOfYear(month, daysInMonth[month])
	if err != nil {
		return 0, 0, err
	}

	return startDay, endDay, nil
}

//region....................................................................................................................... Создание таблиц в БД
//func createTable(db *sql.DB) error {
//	// 1. Создание новой таблицы
//	createTableQuery := `
//		CREATE TABLE tbl_calendar_2025_prod_v2 (
//			day INT PRIMARY KEY,
//			data_state DATE NOT NULL,
//			ves_state INT DEFAULT 0, ves_text TEXT DEFAULT '',
//			hotba_state INT DEFAULT 0,
//			lkarnitin_state INT DEFAULT 0,
//			kofe01_state INT DEFAULT 0,
//			kofe02_state INT DEFAULT 0,
//			kofe03_state INT DEFAULT 0,
//			zub01_state INT DEFAULT 0,
//			zub02_state INT DEFAULT 0,
//			rabotastatus_state INT DEFAULT 0,
//			rabota_state INT DEFAULT 0, rabota_text TEXT DEFAULT '',
//			rabotautro_state INT DEFAULT 0, rabotautro_text TEXT DEFAULT '',
//			rabotaden_state INT DEFAULT 0, rabotaden_text TEXT DEFAULT '',
//			rabotavecher_state INT DEFAULT 0, rabotavecher_text TEXT DEFAULT '',
//			rabotasum_state INT DEFAULT 0, rabotasum_text TEXT DEFAULT '',
//			xxxs_state INT DEFAULT 0,
//			nachalosna_state INT DEFAULT 0, nachalosna_text TEXT DEFAULT '',
//			dlinasna_state INT DEFAULT 0, dlinasna_text TEXT DEFAULT '',
//			krasotasna_state INT DEFAULT 0,
//			xxxm_state INT DEFAULT 0,
//			nastroenieutro_state INT DEFAULT 0,
//			nastroenieden_state INT DEFAULT 0,
//			alkogol_state INT DEFAULT 0,
//			dush_state INT DEFAULT 0,
//			golova_state INT DEFAULT 0,
//			vitamin_state INT DEFAULT 0,
//			fastfood_state INT DEFAULT 0,
//			e1800_state INT DEFAULT 0,
//			muchnoe_state INT DEFAULT 0,
//			klava_state INT DEFAULT 0,
//			maz_state INT DEFAULT 0,
//			tualet_state INT DEFAULT 0
//		);
//	`
//	_, err := db.Exec(createTableQuery)
//	if err != nil {
//		return fmt.Errorf("failed to create new table tbl_my_trackers_new1: %v", err)
//	}
//
//	// 2. Заполнение таблицы 70 строками (с 1 января 2025 по 11 марта 2025)
//	fillDatesQuery := `
//		INSERT INTO tbl_my_trackers_foo (day, data_state, rabota_text)
//		SELECT
//			generate_series(1, 70) AS day,
//			DATE '2025-01-01' + (generate_series(0, 69)) AS data_state,
//			'RbSqF' AS rabota_text;
//	`
//	_, err = db.Exec(fillDatesQuery)
//	if err != nil {
//		return fmt.Errorf("failed to fill dates in tbl_my_trackers_new1: %v", err)
//	}
//
//	return nil
//}
//endregion

// region....................................................................................................................... Копирование таблиц
//func copyTable(db *sql.DB) error {
//	// 1. Создание новой таблицы
//	createTableQuery := `
//		CREATE TABLE tbl_calendar_2025_prod_v2 (
//			day INT PRIMARY KEY,
//			data_state DATE NOT NULL,
//			ves_state INT DEFAULT 0, ves_text TEXT DEFAULT '',
//			hotba_state INT DEFAULT 0,
//			lkarnitin_state INT DEFAULT 0,
//			kofe01_state INT DEFAULT 0,
//			kofe02_state INT DEFAULT 0,
//			kofe03_state INT DEFAULT 0,
//			zub01_state INT DEFAULT 0,
//			zub02_state INT DEFAULT 0,
//			rabotastatus_state INT DEFAULT 0,
//			rabota_state INT DEFAULT 0, rabota_text TEXT DEFAULT '',
//			rabotautro_state INT DEFAULT 0, rabotautro_text TEXT DEFAULT '',
//			rabotaden_state INT DEFAULT 0, rabotaden_text TEXT DEFAULT '',
//			rabotavecher_state INT DEFAULT 0, rabotavecher_text TEXT DEFAULT '',
//			rabotasum_state INT DEFAULT 0, rabotasum_text TEXT DEFAULT '',
//			xxxs_state INT DEFAULT 0,
//			nachalosna_state INT DEFAULT 0, nachalosna_text TEXT DEFAULT '',
//			dlinasna_state INT DEFAULT 0, dlinasna_text TEXT DEFAULT '',
//			krasotasna_state INT DEFAULT 0,
//			xxxm_state INT DEFAULT 0,
//			nastroenieutro_state INT DEFAULT 0,
//			nastroenieden_state INT DEFAULT 0,
//			alkogol_state INT DEFAULT 0,
//			dush_state INT DEFAULT 0,
//			golova_state INT DEFAULT 0,
//			vitamin_state INT DEFAULT 0,
//			fastfood_state INT DEFAULT 0,
//			e1800_state INT DEFAULT 0,
//			muchnoe_state INT DEFAULT 0,
//			klava_state INT DEFAULT 0,
//			maz_state INT DEFAULT 0,
//			tualet_state INT DEFAULT 0
//		);
//	`
//	_, err := db.Exec(createTableQuery)
//	if err != nil {
//		return fmt.Errorf("failed to create new table tbl_calendar_2025_prod_v2: %v", err)
//	}
//
//	// 2. Заполнение таблицы 365 строками (полный 2025 год)
//	fillDatesQuery := `
//		INSERT INTO tbl_calendar_2025_prod_v2 (day, data_state)
//		SELECT
//			generate_series(1, 365) AS day,
//			DATE '2025-01-01' + (generate_series(0, 364)) AS data_state;
//	`
//	_, err = db.Exec(fillDatesQuery)
//	if err != nil {
//		return fmt.Errorf("failed to fill dates in tbl_calendar_2025_prod_v2: %v", err)
//	}
//
//	// 3. Копирование данных из tbl_my_trackers_new с помощью INSERT ... SELECT
//	copyDataQuery := `
//		INSERT INTO tbl_calendar_2025_prod_v2 (
//			day, data_state, ves_state, ves_text, hotba_state, lkarnitin_state,
//			kofe01_state, kofe02_state, kofe03_state, zub01_state, zub02_state,
//			rabotastatus_state, rabota_state, rabota_text, rabotautro_state,
//			rabotautro_text, rabotaden_state, rabotaden_text, rabotavecher_state,
//			rabotavecher_text, rabotasum_state, rabotasum_text, xxxs_state,
//			nachalosna_state, nachalosna_text, dlinasna_state, dlinasna_text,
//			krasotasna_state, xxxm_state, nastroenieutro_state, nastroenieden_state,
//			alkogol_state, dush_state, golova_state, vitamin_state, fastfood_state,
//			e1800_state, muchnoe_state, klava_state, maz_state, tualet_state
//		)
//		SELECT
//			day, data_state, ves_state, ves_text, hotba_state, lkarnitin_state,
//			kofe01_state, kofe02_state, kofe03_state, zub01_state, zub02_state,
//			rabotastatus_state, rabota_state, rabota_text, rabotautro_state,
//			rabotautro_text, rabotaden_state, rabotaden_text, rabotavecher_state,
//			rabotavecher_text, rabotasum_state, rabotasum_text, xxxs_state,
//			nachalosna_state, nachalosna_text, dlinasna_state, dlinasna_text,
//			krasotasna_state, xxxm_state, nastroenieutro_state, nastroenieden_state,
//			alkogol_state, dush_state, golova_state, vitamin_state, fastfood_state,
//			e1800_state, muchnoe_state, klava_state, maz_state, tualet_state
//		FROM tbl_my_trackers_new
//		ON CONFLICT (day) DO UPDATE SET
//			data_state = EXCLUDED.data_state,
//			ves_state = EXCLUDED.ves_state,
//			ves_text = EXCLUDED.ves_text,
//			hotba_state = EXCLUDED.hotba_state,
//			lkarnitin_state = EXCLUDED.lkarnitin_state,
//			kofe01_state = EXCLUDED.kofe01_state,
//			kofe02_state = EXCLUDED.kofe02_state,
//			kofe03_state = EXCLUDED.kofe03_state,
//			zub01_state = EXCLUDED.zub01_state,
//			zub02_state = EXCLUDED.zub02_state,
//			rabotastatus_state = EXCLUDED.rabotastatus_state,
//			rabota_state = EXCLUDED.rabota_state,
//			rabota_text = EXCLUDED.rabota_text,
//			rabotautro_state = EXCLUDED.rabotautro_state,
//			rabotautro_text = EXCLUDED.rabotautro_text,
//			rabotaden_state = EXCLUDED.rabotaden_state,
//			rabotaden_text = EXCLUDED.rabotaden_text,
//			rabotavecher_state = EXCLUDED.rabotavecher_state,
//			rabotavecher_text = EXCLUDED.rabotavecher_text,
//			rabotasum_state = EXCLUDED.rabotasum_state,
//			rabotasum_text = EXCLUDED.rabotasum_text,
//			xxxs_state = EXCLUDED.xxxs_state,
//			nachalosna_state = EXCLUDED.nachalosna_state,
//			nachalosna_text = EXCLUDED.nachalosna_text,
//			dlinasna_state = EXCLUDED.dlinasna_state,
//			dlinasna_text = EXCLUDED.dlinasna_text,
//			krasotasna_state = EXCLUDED.krasotasna_state,
//			xxxm_state = EXCLUDED.xxxm_state,
//			nastroenieutro_state = EXCLUDED.nastroenieutro_state,
//			nastroenieden_state = EXCLUDED.nastroenieden_state,
//			alkogol_state = EXCLUDED.alkogol_state,
//			dush_state = EXCLUDED.dush_state,
//			golova_state = EXCLUDED.golova_state,
//			vitamin_state = EXCLUDED.vitamin_state,
//			fastfood_state = EXCLUDED.fastfood_state,
//			e1800_state = EXCLUDED.e1800_state,
//			muchnoe_state = EXCLUDED.muchnoe_state,
//			klava_state = EXCLUDED.klava_state,
//			maz_state = EXCLUDED.maz_state,
//			tualet_state = EXCLUDED.tualet_state;
//	`
//	_, err = db.Exec(copyDataQuery)
//	if err != nil {
//		return fmt.Errorf("failed to copy data from tbl_my_trackers_new to tbl_calendar_2025_prod_v2: %v", err)
//	}
//
//	return nil
//}
//endregion
