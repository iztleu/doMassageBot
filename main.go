package main

import (
	"database/sql"
	"doMassage/internal/config"
	db2 "doMassage/internal/db"
	"fmt"
	_ "github.com/lib/pq"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
	"log"
	"regexp"
)

var signMenu = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Sign up"),
		tgbotapi.NewKeyboardButton("Cancel")),
)

var mainMenu = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Enroll"),
		tgbotapi.NewKeyboardButton("My schedule")),
)

var massageMenu = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Шейно воротниковый массаж"),
		tgbotapi.NewKeyboardButton("Лечебный массаж")),
)

func main() {
	conf, err := config.LoadConfiguration("config.json")
	if err != nil {
		fmt.Println(err)
	}
	db, err := db2.ConnectingToDb(conf)

	bot, err := tgbotapi.NewBotAPI(conf.TelegramBotToken)
	if err != nil {
		log.Panic(err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = conf.UpdateTimeout

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			UpdateUsername(db, update.Message.From.ID, update.Message.From.UserName)
			if update.Message.IsCommand() {
				cmdText := update.Message.Command()
				if cmdText == "start" {
					if CheckIfUserExists(db, update.Message.From.ID) {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Welcome "+update.Message.From.UserName)
						msg.ReplyMarkup = mainMenu
						bot.Send(msg)
					} else {
						InsertIntoUsers(db, update.Message.From.ID, "", "", "", 0)
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hello, you have to sign up first.")
						msg.ReplyMarkup = signMenu
						bot.Send(msg)
					}
				}
			} else {
				switch update.Message.Text {
				case signMenu.Keyboard[0][0].Text:
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Your fullname:")
					msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
					bot.Send(msg)
					UpdateUserStatus(db, update.Message.From.ID, 1)
				case mainMenu.Keyboard[0][0].Text:
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Choose, what type of massage do you need?")
					msg.ReplyMarkup = massageMenu
					bot.Send(msg)
				case massageMenu.Keyboard[0][0].Text:
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Select date")
					dates := GetDates(db, massageMenu.Keyboard[0][0].Text)
					keyboard := tgbotapi.InlineKeyboardMarkup{}
					for _, date := range dates {
						var row []tgbotapi.InlineKeyboardButton
						btn := tgbotapi.NewInlineKeyboardButtonData(date, date)
						row = append(row, btn)
						keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
					}
					msg.ReplyMarkup = keyboard
					bot.Send(msg)
					InsertIntoBookingList(db, massageMenu.Keyboard[0][0].Text, "", "", update.Message.From.ID, 0)
				case massageMenu.Keyboard[0][1].Text:
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Select date")
					dates := GetDates(db, massageMenu.Keyboard[0][1].Text)
					keyboard := tgbotapi.InlineKeyboardMarkup{}
					for _, date := range dates {
						var row []tgbotapi.InlineKeyboardButton
						btn := tgbotapi.NewInlineKeyboardButtonData(date, date)
						row = append(row, btn)
						keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
					}
					msg.ReplyMarkup = keyboard
					bot.Send(msg)
					InsertIntoBookingList(db, massageMenu.Keyboard[0][1].Text, "", "", update.Message.From.ID, 0)
					fmt.Println("its callback, ", update.CallbackQuery.Data)

				default:
					userStatus := getUserStatus(db, update.Message.From.ID)
					switch userStatus {
					case 1:
						fullname := update.Message.Text
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Your email:")
						msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
						bot.Send(msg)
						UpdateUserStatus(db, update.Message.From.ID, 2)
						UpdateFullname(db, update.Message.From.ID, fullname)
					case 2:
						email := update.Message.Text
						if IsEmailValid(email) {
							UpdateEmail(db, update.Message.From.ID, email)
							UpdateUserStatus(db, update.Message.From.ID, 3)
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Good!")
							msg.ReplyMarkup = mainMenu
							bot.Send(msg)
						} else {
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "invalid email, try again")
							msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
							bot.Send(msg)
							UpdateUserStatus(db, update.Message.From.ID, 1)
						}

					}

				}

			}

		}
		if update.CallbackQuery != nil {
			bookingStatus := GetBookingStatus(db, update.CallbackQuery.From.ID)
			fmt.Println("Status: ", bookingStatus)
			switch bookingStatus {
			case 0:
				date := update.CallbackQuery.Data
				time := GetTime(db, date)
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Select time:")
				keyboard := tgbotapi.InlineKeyboardMarkup{}
				for _, t := range time {
					var row []tgbotapi.InlineKeyboardButton
					btn := tgbotapi.NewInlineKeyboardButtonData(t, t)
					row = append(row, btn)
					keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
				}

				msg.ReplyMarkup = keyboard
				bot.Send(msg)
				UpdateBookingStatus(db, update.CallbackQuery.From.ID, 1)
				UpdateBookingDate(db, update.CallbackQuery.From.ID, date)
			case 1:
				time := update.CallbackQuery.Data
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Good")
				msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
				bot.Send(msg)
				UpdateBookingStatus(db, update.CallbackQuery.From.ID, 2)
				UpdateBookingTime(db, update.CallbackQuery.From.ID, time)
			}

		}
	}

}

func getUserStatus(db *sql.DB, userId int) int {
	var (
		status int
	)
	sqlStatement := "SELECT status FROM users WHERE userId = $1;"
	err := db.QueryRow(sqlStatement, userId).Scan(&status)
	if err != nil {
		log.Fatal("Failed to execute query: ", err)
	}
	return status
}

func UpdateUsername(db *sql.DB, userId int, username string) {
	sqlStatement := `UPDATE users SET username= $2 WHERE userId = $1;`
	_, err := db.Exec(sqlStatement, userId, username)
	if err != nil {
		panic(err)
	}
}
func UpdateFullname(db *sql.DB, userId int, fullname string) {
	sqlStatement := `UPDATE users SET fullname= $2 WHERE userId = $1;`
	_, err := db.Exec(sqlStatement, userId, fullname)
	if err != nil {
		panic(err)
	}
}
func UpdateEmail(db *sql.DB, userId int, email string) {
	sqlStatement := `UPDATE users SET email= $2 WHERE userId = $1;`
	_, err := db.Exec(sqlStatement, userId, email)
	if err != nil {
		panic(err)
	}
}

func UpdateUserStatus(db *sql.DB, userId int, status int) {
	sqlStatement := `UPDATE users SET status = $2 WHERE userId = $1;`
	_, err := db.Exec(sqlStatement, userId, status)
	if err != nil {
		panic(err)
	}
}

func CheckIfUserExists(db *sql.DB, userId int) bool {
	sqlStmt := `SELECT userId FROM users WHERE userId = $1`
	err := db.QueryRow(sqlStmt, userId).Scan(&userId)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Print(err)
		}
		return false
	}
	return true
}

func InsertIntoUsers(db *sql.DB, userId int, fullname string, username string, email string, status int) {
	sqlStmt := `INSERT INTO users (userId, fullName, username, email, status) VALUES ($1, $2, $3, $4, $5)`
	_, err := db.Exec(sqlStmt, userId, fullname, username, email, status)
	if err != nil {
		log.Print(err)
	}
}

func IsEmailValid(e string) bool {
	emailRegex := regexp.MustCompile(`^[A-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	fcbDomain := "@1cb.kz"
	if emailRegex.MatchString(e) && (e[len(e)-7:] == fcbDomain) {
		return true
	}
	return false
}

func GetDates(db *sql.DB, mType string) []string {
	var (
		mDate      string
		datesArray []string
	)
	rows, err := db.Query("select distinct(mDate) from massageSchedule where mType = $1", mType)
	if err != nil {
		fmt.Println(err)
	}
	for rows.Next() {
		err := rows.Scan(&mDate)
		if err != nil {
			fmt.Println(err)
		}
		datesArray = append(datesArray, mDate)
	}
	return datesArray
}
func GetTime(db *sql.DB, date string) []string {
	var (
		mTime     string
		timeArray []string
	)
	rows, err := db.Query("select mTime from massageSchedule where mDate = $1", date)
	if err != nil {
		fmt.Println(err)
	}
	for rows.Next() {
		err := rows.Scan(&mTime)
		if err != nil {
			fmt.Println(err)
		}
		timeArray = append(timeArray, mTime)
	}
	return timeArray
}

func InsertIntoBookingList(db *sql.DB, mType string, mDate string, mTime string, userId int, status int) {
	sqlStmt := `INSERT INTO massageBookingList (mType, mDate, mTime,userId, status) VALUES ($1, $2, $3, $4, $5)`
	_, err := db.Exec(sqlStmt, mType, mDate, mTime, userId, status)
	if err != nil {
		log.Print(err)
	}
}

func GetBookingStatus(db *sql.DB, userId int) int {
	var (
		status int
	)
	sqlStatement := "SELECT distinct(status) FROM massageBookingList WHERE userId = $1;"
	err := db.QueryRow(sqlStatement, userId).Scan(&status)
	if err != nil {
		log.Fatal("Failed to execute query: ", err)
	}
	return status
}
func UpdateBookingStatus(db *sql.DB, userId int, status int) {
	sqlStatement := `UPDATE massageBookingList SET status = $2 WHERE userId = $1;`
	_, err := db.Exec(sqlStatement, userId, status)
	if err != nil {
		panic(err)
	}
}
func UpdateBookingDate(db *sql.DB, userId int, mdate string) {
	sqlStatement := `UPDATE massageBookingList SET mDate= $2 WHERE userId = $1;`
	_, err := db.Exec(sqlStatement, userId, mdate)
	if err != nil {
		panic(err)
	}
}
func UpdateBookingTime(db *sql.DB, userId int, mtime string) {
	sqlStatement := `UPDATE massageBookingList SET mTime= $2 WHERE userId = $1;`
	_, err := db.Exec(sqlStatement, userId, mtime)
	if err != nil {
		panic(err)
	}
}

//func UpdateMassageSchedule(db, )
