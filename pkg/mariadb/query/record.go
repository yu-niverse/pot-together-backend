package query

import (
	"database/sql"
	"errors"
	"fmt"
	"pottogether/pkg/logger"
	"pottogether/pkg/mariadb"
)

type RecordRequest struct {
	UserID       int  `json:"user_id" binding:"required"`
	RoomID       int  `json:"room_id" binding:"required"`
	PotID        int  `json:"pot_id" binding:"required"`
	IngredientID int  `json:"ingredient_id" binding:"required"`
}

type DoneRequest struct {
	Image        string `json:"image" binding:"required"`
	Caption      string `json:"caption" binding:"required"`
	TimeInterval *int   `json:"time_interval" binding:"required"`
	Interrupt    *int   `json:"interrupt" binding:"required"`
	Status       *int   `json:"status" binding:"required"`
}

type RecordOverview struct {
	ID              int     `json:"id"`
	Interval        *int    `json:"interval"`
	Finish_time     *string `json:"finish_time"`
	IngredientID    int     `json:"ingredient_id"`
	IngredientImage string  `json:"ingredient_image"`
	IngredientName  string  `json:"ingredient_name"`
	Status          int     `json:"status"`
}

type RecordDetail struct {
	ID              int     `json:"id"`
	Image           *string `json:"image"`
	Caption         *string `json:"caption"`
	Interval        *int    `json:"interval"`
	Finish_time     *string `json:"finish_time"`
	IngredientID    int     `json:"ingredient_id"`
	IngredientImage string  `json:"ingredient_image"`
	IngredientName  string  `json:"ingredient_name"`
	Interrupt       *int    `json:"interrupt"`
	Status          *int    `json:"status"`
}

func CreateRecord(r RecordRequest) (int, error) {
	// Insert record
	query := "INSERT INTO record (user_id, room_id, pot_id, ingredient_id, time_interval, interrupt, status) VALUES (?, ?, ?, ?, 0, 0, 0)"
	result, err := mariadb.DB.Exec(query, r.UserID, r.RoomID, r.PotID, r.IngredientID)
	if err != nil {
		logger.Error("[RECORD] " + err.Error())
		return -1, err
	}

	// Get record id
	id, err := result.LastInsertId()
	if err != nil {
		logger.Error("[RECORD] " + err.Error())
		return -1, err
	}

	return int(id), nil
}

func GetRecordOverview(id int, condition string) ([]RecordOverview, error) {
	var list []RecordOverview

	query := "SELECT record.id, record.time_interval, record.finish_time, ingredient.id, ingredient.image, ingredient.name, record.status FROM record INNER JOIN ingredient ON record.ingredient_id = ingredient.id WHERE " + condition + " = ?"
	rows, err := mariadb.DB.Query(query, id)
	if err != nil {
		logger.Error("[RECORD] " + err.Error())
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var r RecordOverview
		if err := rows.Scan(&r.ID, &r.Interval, &r.Finish_time, &r.IngredientID, &r.IngredientImage, &r.IngredientName, &r.Status); err != nil {
			logger.Error("[RECORD] " + err.Error())
			return nil, err
		}
		list = append(list, r)
	}
	return list, nil
}

func GetRecordDetail(id int) ([]RecordDetail, error) {
	var list []RecordDetail

	query := "SELECT record.id, record.image, record.caption, record.time_interval, record.finish_time, ingredient.id, ingredient.image, ingredient.name, record.interrupt, record.status FROM record INNER JOIN ingredient ON record.ingredient_id = ingredient.id WHERE record.id = ?"
	rows, err := mariadb.DB.Query(query, id)
	if err != nil {
		logger.Error("[RECORD] " + err.Error())
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var r RecordDetail
		if err := rows.Scan(&r.ID, &r.Image, &r.Caption, &r.Interval, &r.Finish_time, &r.IngredientID, &r.IngredientImage, &r.IngredientName, &r.Interrupt, &r.Status); err != nil {
			logger.Error("[RECORD] " + err.Error())
			return nil, err
		}
		list = append(list, r)
	}
	return list, nil
}

func UpdateRecord(id int, r DoneRequest) error {
	// Check if record exists and status is 0
	query := "SELECT id FROM record WHERE id = ? AND status = 0"
	err := mariadb.DB.QueryRow(query, id).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Warn("[RECORD] Record id:" + fmt.Sprint(id) + " not found or already done")
			return errors.New("record not found or already done")
		} else {
			logger.Error("[RECORD] " + err.Error())
			return err
		}
	}

	// Update record
	query = "UPDATE record SET image = ?, caption = ?, time_interval = ?, interrupt = ?, status = ?, finish_time = NOW() WHERE id = ?"
	_, err = mariadb.DB.Exec(query, r.Image, r.Caption, r.TimeInterval, r.Interrupt, r.Status, id)
	if err != nil {
		logger.Error("[RECORD] " + err.Error())
		return err
	}
	return nil
}
