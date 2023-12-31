package query

import (
	"database/sql"
	"fmt"
	"pottogether/internal/hash"
	"pottogether/pkg/mariadb"
)

type User struct {
	ID       int    `json:"userID"`
	Name     string `json:"name"`
	Avatar   int    `json:"avatar"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserProfile struct {
	ID          int        `json:"userID"`
	Name        string     `json:"name"`
	Avatar      int        `json:"avatar"`
	CookingTime int        `json:"cookingTime"`
	Status      userStatus `json:"status"`
	Done        []string   `json:"done"`
}

type userStatus struct {
	Code       int    `json:"code"`
	Ingredient string `json:"ingredient"`
}

type UserOverview struct {
	ID    int           `json:"userID"`
	Level userLevel     `json:"level"`
	Today []todayRecord `json:"today"`
	Week  []dateRecord  `json:"week"`
	Month []dateRecord  `json:"month"`
}

type userLevel struct {
	Level     int    `json:"level"`
	TotalTime int    `json:"totalTime"`
	Next      string `json:"next"`
}

type todayRecord struct {
	RecordID int    `json:"recordID"`
	Image    string `json:"image"`
}

type dateRecord struct {
	Date   string `json:"date"`
	Length int    `json:"length"`
}

func CheckEmail(email string) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM user WHERE email = ?)"
	var exists bool
	err := mariadb.DB.QueryRow(query, email).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func CheckUser(id int) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM user WHERE id = ?)"
	var exists bool
	err := mariadb.DB.QueryRow(query, id).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func SignUp(user User) (int, error) {
	var err error
	// Hash password
	if user.Password, err = hash.HashPassword(user.Password); err != nil {
		return -1, err
	}
	// Insert into user database
	query := `
		INSERT INTO user (avatar, email, username, password, created_at, level, total_time) 
		VALUES (?, ?, ?, ?, NOW(), 1, 0)`
	result, err := mariadb.DB.Exec(query, user.Avatar, user.Email, user.Name, user.Password)
	if err != nil {
		return -1, err
	}
	// Get auto increment id
	id, err := result.LastInsertId()
	if err != nil {
		return -1, err
	}
	return int(id), nil
}

func Login(email string, input_pwd string) (int, error) {
	// Get user password
	var id int
	var password string
	query := "SELECT id, password FROM user WHERE email = ?"
	err := mariadb.DB.QueryRow(query, email).Scan(&id, &password)
	if err != nil {
		if err == sql.ErrNoRows {
			return -1, nil
		}
		return -1, err
	}
	// Check if password is correct
	err = hash.CheckPasswordHash(input_pwd, password)
	if err != nil {
		return -1, nil
	}
	return id, nil
}

func GetProfile(id int) (UserProfile, error) {
	var result UserProfile
	// Get user info
	query := "SELECT id, avatar, username FROM user WHERE id = ?"
	err := mariadb.DB.QueryRow(query, id).Scan(&result.ID, &result.Avatar, &result.Name)
	if err != nil {
		return result, err
	}
	// Get cooking time
	query = `
		SELECT TIMESTAMPDIFF(SECOND, created_at, NOW()) FROM record
		WHERE user_id = ? AND status = 0
		ORDER BY created_at DESC LIMIT 1`
	err = mariadb.DB.QueryRow(query, id).Scan(&result.CookingTime)
	if err != nil {
		if err == sql.ErrNoRows {
			result.CookingTime = 0
		} else {
			return result, err
		}
	}
	// Get status
	query = `
		SELECT status, ingredient.name FROM record
		INNER JOIN ingredient ON record.ingredient_id = ingredient.id
		WHERE user_id = ?
		ORDER BY created_at DESC LIMIT 1`
	err = mariadb.DB.QueryRow(query, id).Scan(&result.Status.Code, &result.Status.Ingredient)
	if err != nil {
		if err == sql.ErrNoRows {
			result.Status.Code = 0
			result.Status.Ingredient = ""
		} else {
			return result, err
		}
	}
	// Get done
	query = `
		SELECT ingredient.name FROM record
		INNER JOIN ingredient ON record.ingredient_id = ingredient.id
		WHERE user_id = ? AND status = 1
		ORDER BY created_at DESC LIMIT 5`
	rows, err := mariadb.DB.Query(query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return result, nil
		}
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			return result, err
		}
		result.Done = append(result.Done, name)
	}
	return result, nil
}

func GetOverview(id int) (UserOverview, error) {
	var result UserOverview
	// Get user info
	query := "SELECT id, level, total_time FROM user WHERE id = ?"
	err := mariadb.DB.QueryRow(query, id).Scan(&result.ID, &result.Level.Level, &result.Level.TotalTime)
	if err != nil {
		if err == sql.ErrNoRows {
			return result, fmt.Errorf("user %d does not exist", id)
		}
		return result, err
	}
	// Get next level
	result.Level.Next, err = getNextLevel(result.Level.Level)
	// Get today
	result.Today, err = getToday(id)
	if err != nil {
		return result, err
	}
	// Get week
	result.Week, err = getWeekInterval(id)
	if err != nil {
		return result, err
	}
	// Get month
	result.Month, err = getMonthInterval(id)
	if err != nil {
		return result, err
	}
	return result, nil
}

func getToday(id int) ([]todayRecord, error) {
	var records []todayRecord
	query := `
		SELECT record.id, ingredient.image 
		FROM record INNER JOIN ingredient 
		ON record.ingredient_id = ingredient.id 
		WHERE user_id = ? AND DATE(finish_time) = CURDATE()`
	rows, err := mariadb.DB.Query(query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return records, nil
		}
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var record todayRecord
		err = rows.Scan(&record.RecordID, &record.Image)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

func getWeekInterval(id int) ([]dateRecord, error) {
	var records []dateRecord
	query := `
		SELECT 
			DATE(created_at) AS date,
			SUM(time_interval) AS total_time
		FROM record
		WHERE WEEK(created_at) = WEEK(NOW()) AND user_id = ?
		GROUP BY DATE(created_at);`
	rows, err := mariadb.DB.Query(query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return records, nil
		}
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var record dateRecord
		err = rows.Scan(&record.Date, &record.Length)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

func getMonthInterval(id int) ([]dateRecord, error) {
	var records []dateRecord
	query := `
		SELECT 
			DATE(created_at) AS date,
			SUM(time_interval) AS total_time
		FROM record
		WHERE MONTH(created_at) = MONTH(NOW()) AND user_id = ?
		GROUP BY DATE(created_at);`
	rows, err := mariadb.DB.Query(query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return records, nil
		}
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var record dateRecord
		err = rows.Scan(&record.Date, &record.Length)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

func getNextLevel(level int) (string, error) {
	query := `
		SELECT image FROM ingredient
		WHERE requirement = "level%d"`
	query = fmt.Sprintf(query, level+1)
	var image string
	err := mariadb.DB.QueryRow(query).Scan(&image)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return image, nil
}
