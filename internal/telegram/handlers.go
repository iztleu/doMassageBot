package main

import (
	"database/sql"
	"doMassageBot/internal/config"
	db2 "doMassageBot/internal/db"
	"fmt"
	_ "github.com/lib/pq"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
	"log"
	"net/mail"
)

var signMenu = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Sign up"),
		tgbotapi.NewKeyboardButton("Cancel")),
)

var mainMenu = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Enroll"),
		tgbotapi.NewKeyboardButton("Rating")),
)

var massageMenu = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Шейно воротниковый массаж"),
		tgbotapi.NewKeyboardButton("Лечебный массаж")),
)

type Users struct {
	State    int
	Fullname string
	Username string
	Email    string
}

var usersMap map[int]*Users

func init() {
	usersMap = make(map[int]*Users)
}

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
		log.Printf("[%s] [%s] %s", update.Message.From.ID, update.Message.From.UserName, update.Message.Text)
		if update.Message != nil {
			if CheckIfUserExists(db, update.Message.From.ID) {
				_ = tgbotapi.NewMessage(update.Message.Chat.ID, "Select date:")
				classes := [3]string{"28.04", "29.04", "30.04"}
				keyboard := tgbotapi.InlineKeyboardMarkup{}
				for _, class := range classes {
					var row []tgbotapi.InlineKeyboardButton
					btn := tgbotapi.NewInlineKeyboardButtonData(class, class)
					row = append(row, btn)
					keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
				}
			}

			//	msg.ReplyMarkup = keyboard
			//	bot.Send(msg)
			//	fmt.Println(update.CallbackQuery)
			//	//switch update.Message.Text {
			//	//case "28.04":
			//	//	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Select time:")
			//	//	classes := []string{"9:00", "10:00", "11:00"}
			//	//	keyboard := tgbotapi.InlineKeyboardMarkup{}
			//	//	for _, class := range classes {
			//	//		var row []tgbotapi.InlineKeyboardButton
			//	//		btn := tgbotapi.NewInlineKeyboardButtonData(class, class)
			//	//		row = append(row, btn)
			//	//		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
			//	//	}
			//	//	msg.ReplyMarkup = keyboard
			//	//	bot.Send(msg)
			//	//}
			//}

			if CheckIfUserExists(db, update.Message.From.ID) {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Welcome "+update.Message.From.UserName)
				msg.ReplyMarkup = mainMenu
				bot.Send(msg)
				if update.Message.Text == mainMenu.Keyboard[0][0].Text {

					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Choose, what type of massage do you need?")
					msg.ReplyMarkup = massageMenu
					bot.Send(msg)
					if update.Message.Text == massageMenu.Keyboard[0][1].Text {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Select date")
						classes := []string{"cm", "dkmk", "dvns"}
						keyboard := tgbotapi.InlineKeyboardMarkup{}
						for _, class := range classes {
							var row []tgbotapi.InlineKeyboardButton
							btn := tgbotapi.NewInlineKeyboardButtonData(class, class)
							row = append(row, btn)
							keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
						}

						msg.ReplyMarkup = keyboard
						bot.Send(msg)
					}

				}

			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hello, you have to sign up first.")
				msg.ReplyMarkup = signMenu
				bot.Send(msg)
				if update.Message.Text == signMenu.Keyboard[0][0].Text {
					usersMap[update.Message.From.ID] = new(Users)
					usersMap[update.Message.From.ID].State = 0
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Your fullname:")
					msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
					bot.Send(msg)
				} else {
					u, ok := usersMap[update.Message.From.ID]
					fmt.Println(" +++", u, ok)
					if ok {
						if u.State == 0 {
							u.Fullname = update.Message.Text
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Your email:")
							msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
							bot.Send(msg)
							u.State = 1
						} else if u.State == 1 {
							u.Email = update.Message.Text
							if IsValidEmail(u.Email) {
								msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Good!")
								msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
								bot.Send(msg)
								u.Username = update.Message.From.UserName
								InsertIntoUsers(db, update.Message.From.ID, u.Fullname, u.Username, u.Email)
							} else {
								msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Invalid email")
								msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
								bot.Send(msg)
							}

						}
					}
				}

			}

		}

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
func InsertIntoUsers(db *sql.DB, userId int, fullname string, username string, email string) {
	sqlStmt := `INSERT INTO users (userId, fullName, username, email) VALUES ($1, $2, $3, $4)`
	_, err := db.Exec(sqlStmt, userId, fullname, username, email)
	if err != nil {
		log.Print(err)
	}
}

func IsValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
