package db

import (
	"database/sql"
	entity "doMassage/internal/entity"
	"fmt"
	"log"
	"regexp"
)

func GetUserStatus(db *sql.DB, userId int) int {
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
func UpdateScheduleTime(db *sql.DB, userId int, mtime string) {
	sqlStatement := `UPDATE massageSchedule SET mTime= $2 WHERE userId = $1;`
	_, err := db.Exec(sqlStatement, userId, mtime)
	if err != nil {
		panic(err)
	}
}

func GetMySchedule(db *sql.DB, userId int) []entity.MySchedule {
	var (
		obj      entity.MySchedule
		mDate    string
		mTime    string
		mType    string
		objArray []entity.MySchedule
	)
	rows, err := db.Query("SELECT m.mDate,m.mTime, m.mType FROM massageBookingList as m JOIN users as u ON u.userId = $1;", userId)

	if err != nil {
		log.Fatal("Failed to execute query: ", err)
	}

	for rows.Next() {
		err := rows.Scan(&mDate, &mTime, &mType)
		if err != nil {
			log.Fatal("Failed to execute query: ", err)
		}
		obj.MType = mType
		obj.MDate = mDate
		obj.MTime = mTime

		objArray = append(objArray, obj)

	}
	return objArray

}
func GenerateTime(db *sql.DB) []string {
	var (
		hours      string
		hoursArray []string
	)
	rows, err := db.Query("SELECT  x::time from generate_series('2021-01-01 09:00:00','2021-01-01 17:00:00',INTERVAL '30 minutes')t(x)")

	if err != nil {
		log.Fatal("Failed to execute query: ", err)
	}
	for rows.Next() {
		err := rows.Scan(&hours)
		if err != nil {
			log.Fatal("Failed to execute query: ", err)
		}
		hoursArray = append(hoursArray, hours)

	}
	return hoursArray

}

func GetScheduleStatus(db *sql.DB, userId int) int {
	var (
		status int
	)
	sqlStatement := "SELECT distinct(status) FROM massageSchedule WHERE userId = $1;"
	err := db.QueryRow(sqlStatement, userId).Scan(&status)
	if err != nil {
		log.Fatal("Failed to execute query: ", err)
	}
	return status
}

func InsertIntoSchedule(db *sql.DB, mType string, mDate string, mTime string, userId int, status int) {
	if mDate == "Today" {
		sqlStmt := `INSERT INTO massageSchedule(mid, mDate, mTime, uId,status) VALUES((SELECT id from massageType WHERE mType=$1),CURRENT_DATE,$2, (SELECT id from users WHERE userId= $3));)`
		_, err := db.Exec(sqlStmt, mType, mTime, userId, status)
		if err != nil {
			log.Print(err)
		}

	}
}
func UpdateScheduleMType(db *sql.DB, userId int, mDate string, status int) {
	sqlStatement := `UPDATE massageSchedule SET m, status = $3 WHERE userId = $1;`
	_, err := db.Exec(sqlStatement, userId, mDate, status)
	if err != nil {
		panic(err)
	}
}
