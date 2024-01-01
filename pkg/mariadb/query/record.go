package query

import (
	"pottogether/pkg/logger"
	"pottogether/pkg/mariadb"
	"strconv"
)

type Record struct {
	ID           int    `json:"recordID"`
	UserID       int    `json:"userID"`
	RoomID       int    `json:"roomID"`
	PotID        string `json:"potID"`
	IngredientID int    `json:"ingredientID"`
	Image        string `json:"image"`
	Caption      string `json:"caption"`
	Interval     int    `json:"interval"`
	FinishTime   int    `json:"finishTime"`
	Interrupt    int    `json:"interrupt"`
	Status       int    `json:"status"`
}

type RecordDetail struct {
	ID              int    `json:"recordID"`
	Username        string `json:"username"`
	Image           string `json:"image"`
	Caption         string `json:"caption"`
	Interval        int    `json:"interval"`
	FinishTime      int    `json:"finishTime"`
	IngredientID    int    `json:"ingredientID"`
	IngredientImage string `json:"ingredientImage"`
	IngredientName  string `json:"ingredientName"`
	Interrupt       int    `json:"interrupt"`
	Status          int    `json:"status"`
}

func CreateRecord(record Record) (int, error) {
	query := `
		INSERT INTO record (user_id, room_id, pot_id, ingredient_id, time_interval, interrupt, status, created_at, finish_time, image, caption)
		VALUES (?, ?, ?, ?, 0, 0, 0, NOW(), NOW(), "null", "null")`
	result, err := mariadb.DB.Exec(query, record.UserID, record.RoomID, record.PotID, record.IngredientID)
	if err != nil {
		return -1, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return -1, err
	}
	return int(id), nil
}

func UpdateRecord(record Record) error {
	// check if record exists
	query := `SELECT id FROM record WHERE id = ?`
	err := mariadb.DB.QueryRow(query, record.ID).Scan(&record.ID)
	if err != nil {
		logger.Warn("Invalid recordID: " + strconv.Itoa(record.ID))
		return err
	}
	// update record
	query = `
		UPDATE record
		SET image = ?, caption = ?, time_interval = ?, finish_time = NOW(), interrupt = ?, status = ?
		WHERE id = ?
	`
	_, err = mariadb.DB.Exec(query, record.Image, record.Caption, record.Interval, record.Interrupt, record.Status, record.ID)
	if err != nil {
		return err
	}
	return nil
}

func GetUserRecords(userID int) ([]RecordDetail, error) {
	query := `
		SELECT r.id, r.image, r.caption, r.time_interval, UNIX_TIMESTAMP(r.finish_time), r.ingredient_id, i.name, i.image, r.interrupt, r.status, u.username
		FROM record r
		INNER JOIN ingredient i ON r.ingredient_id = i.id
		INNER JOIN user u ON r.user_id = u.id
		WHERE r.user_id = ?
		ORDER BY status DESC`
	rows, err := mariadb.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var records []RecordDetail
	for rows.Next() {
		var record RecordDetail
		err = rows.Scan(&record.ID, &record.Image, &record.Caption, &record.Interval, &record.FinishTime, &record.IngredientID, &record.IngredientName, &record.IngredientImage, &record.Interrupt, &record.Status, &record.Username)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

func GetRecordDetail(recordID int) (RecordDetail, error) {
	// check if record exists
	query := `SELECT id FROM record WHERE id = ?`
	err := mariadb.DB.QueryRow(query, recordID).Scan(&recordID)
	if err != nil {
		logger.Warn("Invalid recordID: " + strconv.Itoa(recordID))
		return RecordDetail{}, err
	}
	query = `
		SELECT r.id, r.image, r.caption, r.time_interval, UNIX_TIMESTAMP(r.finish_time), r.ingredient_id, i.name, i.image, r.interrupt, r.status, u.username
		FROM record r
		INNER JOIN ingredient i ON r.ingredient_id = i.id
		INNER JOIN user u ON r.user_id = u.id
		WHERE r.id = ?`
	var record RecordDetail
	err = mariadb.DB.QueryRow(query, recordID).Scan(&record.ID, &record.Image, &record.Caption, &record.Interval, &record.FinishTime, &record.IngredientID, &record.IngredientName, &record.IngredientImage, &record.Interrupt, &record.Status, &record.Username)
	if err != nil {
		return RecordDetail{}, err
	}
	return record, nil
}

func GetRoomRecords(roomID int) ([]RecordDetail, error) {
	// check if room exists
	query := `SELECT id FROM room WHERE id = ?`
	err := mariadb.DB.QueryRow(query, roomID).Scan(&roomID)
	if err != nil {
		logger.Warn("Invalid roomID: " + strconv.Itoa(roomID))
		return nil, err
	}
	query = `
		SELECT r.id, r.image, r.caption, r.time_interval, UNIX_TIMESTAMP(r.finish_time), r.ingredient_id, i.name, i.image, r.interrupt, r.status, u.username
		FROM record r
		INNER JOIN ingredient i ON r.ingredient_id = i.id
		INNER JOIN user u ON r.user_id = u.id
		WHERE r.room_id = ?
		ORDER BY status DESC`
	rows, err := mariadb.DB.Query(query, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var records []RecordDetail
	for rows.Next() {
		var record RecordDetail
		err = rows.Scan(&record.ID, &record.Image, &record.Caption, &record.Interval, &record.FinishTime, &record.IngredientID, &record.IngredientName, &record.IngredientImage, &record.Interrupt, &record.Status, &record.Username)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}
