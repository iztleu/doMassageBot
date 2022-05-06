package main

import (
	"doMassageBot/internal/config"
	db2 "doMassageBot/internal/db"
	query "doMassageBot/internal/db"
	"fmt"
	_ "github.com/lib/pq"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
	"log"
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

var TimeMenu = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Today"),
		tgbotapi.NewKeyboardButton("Tomorrow")),
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
			query.UpdateUsername(db, update.Message.From.ID, update.Message.From.UserName)
			if update.Message.IsCommand() {
				cmdText := update.Message.Command()
				if cmdText == "start" {
					if query.CheckIfUserExists(db, update.Message.From.ID) {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Welcome "+update.Message.From.UserName)
						msg.ReplyMarkup = mainMenu
						bot.Send(msg)
					} else {
						query.InsertIntoUsers(db, update.Message.From.ID, "", "", "", 0)
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
					query.UpdateUserStatus(db, update.Message.From.ID, 1)
				case mainMenu.Keyboard[0][0].Text:
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Choose, When?")
					msg.ReplyMarkup = TimeMenu
					bot.Send(msg)
					//msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Choose, what type of massage do you need?")
					//msg.ReplyMarkup = massageMenu
					//bot.Send(msg)

				case TimeMenu.Keyboard[0][0].Text:
					query.InsertIntoSchedule(db, "", TimeMenu.Keyboard[0][0].Text, "", update.Message.From.ID, 0)
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Choose, what type of massage do you need?")
					msg.ReplyMarkup = massageMenu
					bot.Send(msg)
				case mainMenu.Keyboard[0][1].Text:
					objs := query.GetMySchedule(db, update.Message.From.ID)
					if len(objs) > 0 {
						for _, obj := range objs {
							text := fmt.Sprintf("*%s*\n"+"*Date: * _%v_\n"+"*Time: * _%s_\n", obj.MType, obj.MDate, obj.MTime)
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
							msg.ParseMode = "markdown"
							bot.Send(msg)
						}
					} else {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "You didn't enrolled yet...")
						msg.ReplyMarkup = mainMenu
						bot.Send(msg)

					}
				case massageMenu.Keyboard[0][0].Text:

					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Select time")
					timeArray := query.GenerateTime(db)
					keyboard := tgbotapi.InlineKeyboardMarkup{}
					for _, time := range timeArray {
						var row []tgbotapi.InlineKeyboardButton
						btn := tgbotapi.NewInlineKeyboardButtonData(time[11:len(time)-4], time[11:len(time)-4])
						row = append(row, btn)
						keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
					}
					msg.ReplyMarkup = keyboard
					bot.Send(msg)
					//query.InsertIntoBookingList(db, massageMenu.Keyboard[0][0].Text, "", "", update.Message.From.ID, 0)
				case massageMenu.Keyboard[0][1].Text:
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Select date")
					dates := query.GetDates(db, massageMenu.Keyboard[0][1].Text)
					keyboard := tgbotapi.InlineKeyboardMarkup{}
					for _, date := range dates {
						var row []tgbotapi.InlineKeyboardButton
						btn := tgbotapi.NewInlineKeyboardButtonData(date, date)
						row = append(row, btn)
						keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
					}
					msg.ReplyMarkup = keyboard
					bot.Send(msg)
					//query.InsertIntoBookingList(db, massageMenu.Keyboard[0][1].Text, "", "", update.Message.From.ID, 0)
					//fmt.Println("its callback, ", update.CallbackQuery.Data)

				default:
					userStatus := query.GetUserStatus(db, update.Message.From.ID)
					switch userStatus {
					case 1:
						fullname := update.Message.Text
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Your email:")
						msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
						bot.Send(msg)
						query.UpdateUserStatus(db, update.Message.From.ID, 2)
						query.UpdateFullname(db, update.Message.From.ID, fullname)
					case 2:
						email := update.Message.Text
						if query.IsEmailValid(email) {
							query.UpdateEmail(db, update.Message.From.ID, email)
							query.UpdateUserStatus(db, update.Message.From.ID, 3)
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "You successfully signed up.")
							msg.ReplyMarkup = mainMenu
							bot.Send(msg)
						} else {
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "invalid email, try again")
							msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
							bot.Send(msg)
							query.UpdateUserStatus(db, update.Message.From.ID, 1)
						}

					}

				}

			}

		}
		if update.CallbackQuery != nil {
			bookingStatus := query.GetScheduleStatus(db, update.CallbackQuery.From.ID)
			fmt.Println("Status:", bookingStatus)
			switch bookingStatus {
			case 0:
				date := update.CallbackQuery.Data
				time := query.GetTime(db, date)
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
				query.UpdateBookingStatus(db, update.CallbackQuery.From.ID, 1)
				query.UpdateBookingDate(db, update.CallbackQuery.From.ID, date)
				//case 1:
				//	time := update.CallbackQuery.Data
				//	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Good")
				//	msg.ReplyMarkup = mainMenu
				//	bot.Send(msg)
				//	query.UpdateBookingStatus(db, update.CallbackQuery.From.ID, 2)
				//	//query.UpdateBookingTime(db, update.CallbackQuery.From.ID, time)
			}

		}

	}
}
