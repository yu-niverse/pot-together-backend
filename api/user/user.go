package user

import (
	"errors"
	"pottogether/pkg/logger"
	"pottogether/pkg/mariadb"

	"golang.org/x/crypto/bcrypt"
)

type UserInfo struct {
	ID        int           `json:"id"`
	Avatar    *int          `json:"avatar"`
	Name      string        `json:"name"`
	Level     int           `json:"level"`
	TotalTime int           `json:"total_time"`
	Today     []todayRecord `json:"today"`
	Week      []int         `json:"week"`
}

type todayRecord struct {
	ID    int    `json:"id"`
	Image string `json:"image"`
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err
}

func signUp(rr signUpRequest) (int, error) {
	var email string

	// Check if user already exists
	query := "SELECT email FROM user WHERE email = ?"
	err := mariadb.DB.QueryRow(query, rr.Email).Scan(&email)
	if err != nil && err.Error() != "sql: no rows in result set" {
		logger.Error("[USER] " + err.Error())
		return -1, err
	} else if email != "" {
		logger.Warn("[USER] Email:" + rr.Email + " already exists")
		return -1, errors.New("email already exists")
	}

	// Hash password
	if rr.Passwd, err = hashPassword(rr.Passwd); err != nil {
		logger.Error("[USER] " + err.Error())
		return -1, err
	}

	// Insert into user database
	query = "INSERT INTO user (avatar, email, username, password, created_at, level) VALUES (?, ?, ?, ?, NOW(), 1)"
	result, err := mariadb.DB.Exec(query, rr.Avatar, rr.Email, rr.Name, rr.Passwd)
	if err != nil {
		logger.Error("[USER] " + err.Error())
		return -1, err
	}

	// Get auto incremented id
	id, err := result.LastInsertId()
	if err != nil {
		logger.Error("[USER] " + err.Error())
		return -1, err
	}

	logger.Info("[USER] Successfully registered user with email: " + rr.Email)
	return int(id), nil
}

func login(lr loginRequest) (int, error) {
	var password string
	var id int

	// Get user password
	query := "SELECT id, password FROM user WHERE email = ?"
	err := mariadb.DB.QueryRow(query, lr.Email).Scan(&id, &password)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			logger.Warn("[USER] Email: " + lr.Email + " not found")
			return -1, errors.New("user not found")
		}
		logger.Error("[USER] " + err.Error())
		return -1, err
	}

	// Check if password is correct
	err = checkPasswordHash(lr.Passwd, password)
	if err != nil {
		logger.Warn("[USER] Incorrect password for Email: " + lr.Email)
		return -1, errors.New("incorrect password")
	}

	logger.Info("[USER] Successfully logged in user with email: " + lr.Email)

	return id, nil
}

func getProfile(id int) (UserInfo, error) {
	var ui UserInfo

	// Get user info
	query := "SELECT id, avatar, username, level FROM user WHERE id = ?"
	err := mariadb.DB.QueryRow(query, id).Scan(&ui.ID, &ui.Avatar, &ui.Name, &ui.Level)
	if err != nil {
		logger.Error("[USER] " + err.Error())
		return ui, err
	}

	// Get total time
	query = "SELECT COALESCE(SUM(time_interval), 0) FROM record WHERE user_id = ?"
	err = mariadb.DB.QueryRow(query, id).Scan(&ui.TotalTime)
	if err != nil {
		logger.Error("[USER] " + err.Error())
		return ui, err
	}

	// Get today list
	ui.Today, err = getToday(id)
	if err != nil {
		logger.Error("[USER] " + err.Error())
		return ui, err
	}

	// Get week interval
	ui.Week, err = getWeekInterval(id)
	if err != nil {
		logger.Error("[USER] " + err.Error())
		return ui, err
	}

	return ui, nil
}

func getToday(id int) ([]todayRecord, error) {
	var todayList []todayRecord

	// Get user record
	query := "SELECT record.id, ingredient.image FROM record INNER JOIN ingredient ON record.ingredient_id = ingredient.id WHERE user_id = ? AND DATE(finish_time) = CURDATE()"
	rows, err := mariadb.DB.Query(query, id)
	if err != nil {
		logger.Error("[USER] " + err.Error())
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var t todayRecord
		err = rows.Scan(&t.ID, &t.Image)
		if err != nil {
			logger.Error("[USER] " + err.Error())
			return nil, err
		}
		todayList = append(todayList, t)
	}
	return todayList, nil
}

func getWeekInterval(id int) ([]int, error) {
	var weekInterval []int

	// Go through 7 days and sum up the time_interval
	for i := 6; i >= 0; i-- {
		query := "SELECT COALESCE(SUM(time_interval), 0) FROM record WHERE user_id = ? AND DATE(finish_time) = DATE_SUB(CURDATE(), INTERVAL ? DAY)"
		var sum int
		err := mariadb.DB.QueryRow(query, id, i).Scan(&sum)
		if err != nil {
			logger.Error("[USER] " + err.Error())
			return weekInterval, err
		}
		weekInterval = append(weekInterval, sum)
	}
	return weekInterval, nil
}
