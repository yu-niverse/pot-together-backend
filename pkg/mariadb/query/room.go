package query

import (
	"database/sql"
	"fmt"
	"pottogether/pkg/mariadb"
	"strings"

	"github.com/google/uuid"
)

type Room struct {
	ID          int    `json:"roomID"`
	Name        string `json:"name"`
	MemberLimit int    `json:"memberLimit"`
	Privacy     string `json:"privacy"`
	Category    string `json:"category"`
}

type RoomDetail struct {
	ID          int      `json:"roomID"`
	Name        string   `json:"name"`
	MemberCnt   int      `json:"memberCnt"`
	MemberLimit int      `json:"memberLimit"`
	Category    []string `json:"category"`
}

type RoomOverview struct {
	ID         int              `json:"roomID"`
	CurrentPot string           `json:"currentPot"`
	Name       string           `json:"name"`
	Members    []roomUser       `json:"members"`
	Week       []roomDateRecord `json:"week"`
	Level      userLevel        `json:"level"`
	Cooking    []todayRecord    `json:"cooking"`
	Done       []todayRecord    `json:"done"`
}

type roomUser struct {
	ID     int  `json:"userID"`
	Avatar *int `json:"avatar"`
}

type roomDateRecord struct {
	Date      string `json:"date"`
	UserTotal int    `json:"userTotal"`
	RoomTotal int    `json:"roomTotal"`
}

func CreateRoom(room Room, userID int) (int, string, error) {
	// begin transaction
	tx, err := mariadb.DB.Begin()
	if err != nil {
		return -1, "", err
	}
	// generate pot uuid
	potID := uuid.NewString()
	query := `
		INSERT INTO room (roomname, current_pot, member_cnt, member_limit, privacy, category, level, total_time, created_at) 
		VALUES (?, ?, ?, ?, ?, ?, 1, 0, NOW())`
	result, err := tx.Exec(query, room.Name, potID, 1, room.MemberLimit, room.Privacy, room.Category)
	if err != nil {
		tx.Rollback()
		return -1, "", err
	}
	// Get auto increment id
	id, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return -1, "", err
	}
	// create pot
	query = `
		INSERT INTO pot (id, room_id)
		VALUES (?, ?)`
	_, err = tx.Exec(query, potID, id)
	if err != nil {
		tx.Rollback()
		return -1, "", err
	}
	// add user to room
	query = `
		INSERT INTO room_user (user_id, room_id)
		VALUES (?, ?)`
	_, err = tx.Exec(query, userID, id)
	if err != nil {
		tx.Rollback()
		return -1, "", err
	}
	// commit transaction
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return -1, "", err
	}
	return int(id), potID, nil
}

func GetRooms(userID int) ([]RoomDetail, error) {
	rooms := []RoomDetail{}
	query := `
		SELECT r.id, r.roomname, r.member_cnt, r.member_limit, r.category
		FROM room r
		INNER JOIN room_user ru ON r.id = ru.room_id
		WHERE ru.user_id = ?`
	rows, err := mariadb.DB.Query(query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return rooms, nil
		}
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var room RoomDetail
		var category string
		if err := rows.Scan(&room.ID, &room.Name, &room.MemberCnt, &room.MemberLimit, &category); err != nil {
			return nil, err
		}
		room.Category = strings.Split(category, "|")
		rooms = append(rooms, room)
	}
	return rooms, nil
}

func GetPublicRooms() ([]RoomDetail, error) {
	rooms := []RoomDetail{}
	query := `
		SELECT r.id, r.roomname, r.member_cnt, r.member_limit, r.category
		FROM room r
		WHERE r.privacy = 'public'`
	rows, err := mariadb.DB.Query(query)
	if err != nil {
		if err == sql.ErrNoRows {
			return rooms, nil
		}
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var room RoomDetail
		var category string
		if err := rows.Scan(&room.ID, &room.Name, &room.MemberCnt, &room.MemberLimit, &category); err != nil {
			return nil, err
		}
		room.Category = strings.Split(category, "|")
		rooms = append(rooms, room)
	}
	return rooms, nil
}

func GetRoomOverview(roomID int, userID int) (RoomOverview, error) {
	// Check if room exists
	query := "SELECT EXISTS(SELECT 1 FROM room WHERE id = ?)"
	var exists bool
	err := mariadb.DB.QueryRow(query, roomID).Scan(&exists)
	if err != nil {
		return RoomOverview{}, err
	} else if !exists {
		return RoomOverview{}, fmt.Errorf("room does not exist")
	}
	var room RoomOverview
	// Get room info
	query = `
		SELECT r.id, r.roomname, r.current_pot, r.level, r.total_time
		FROM room r WHERE r.id = ?`
	err = mariadb.DB.QueryRow(query, roomID).Scan(&room.ID, &room.Name, &room.CurrentPot, &room.Level.Level, &room.Level.TotalTime)
	if err != nil {
		return room, err
	}
	// Get next level
	room.Level.Next, err = getNextLevel(room.Level.Level)
	// Get room members
	query = `
		SELECT u.id, u.avatar
		FROM room_user ru
		INNER JOIN user u ON ru.user_id = u.id
		WHERE ru.room_id = ?`
	rows, err := mariadb.DB.Query(query, roomID)
	if err != nil {
		return room, err
	}
	defer rows.Close()
	for rows.Next() {
		var member roomUser
		if err := rows.Scan(&member.ID, &member.Avatar); err != nil {
			return room, err
		}
		room.Members = append(room.Members, member)
	}
	// Get week interval
	room.Week, err = getRoomWeekInterval(roomID, userID)
	if err != nil {
		return room, err
	}
	// Get cooking records
	room.Cooking, err = getCookingRecords(roomID)
	if err != nil {
		return room, err
	}
	// Get done records
	room.Done, err = getDoneRecords(roomID)
	if err != nil {
		return room, err
	}
	return room, nil
}

func getRoomWeekInterval(roomID int, userID int) ([]roomDateRecord, error) {
	weekInterval := []roomDateRecord{}
	// Get room records
	query := `
		SELECT
			DATE(created_at) AS date,
			SUM(time_interval) AS total_time
		FROM record
		WHERE WEEK(created_at) = WEEK(NOW()) AND room_id = ?
		GROUP BY date`
	rows, err := mariadb.DB.Query(query, roomID)
	if err != nil {
		if err == sql.ErrNoRows {
			return weekInterval, nil
		}
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var record roomDateRecord
		if err := rows.Scan(&record.Date, &record.RoomTotal); err != nil {
			return weekInterval, err
		}
		weekInterval = append(weekInterval, record)
	}
	// Get user records
	query = `
		SELECT
			DATE(created_at) AS date,
			SUM(time_interval) AS total_time
		FROM record
		WHERE WEEK(created_at) = WEEK(NOW()) AND user_id = ? AND room_id = ?
		GROUP BY date`
	rows, err = mariadb.DB.Query(query, userID, roomID)
	if err != nil && err != sql.ErrNoRows {
		return weekInterval, err
	}
	defer rows.Close()
	for rows.Next() {
		var record roomDateRecord
		if err := rows.Scan(&record.Date, &record.UserTotal); err != nil {
			return weekInterval, err
		}
		for i, r := range weekInterval {
			if r.Date == record.Date {
				weekInterval[i].UserTotal = record.UserTotal
				break
			}
		}
	}
	return weekInterval, nil
}

func getCookingRecords(roomID int) ([]todayRecord, error) {
	records := []todayRecord{}
	query := `
		SELECT r.id, i.image
		FROM record r
		INNER JOIN ingredient i
		ON r.ingredient_id = i.id
		WHERE r.room_id = ? AND r.status = 0
		ORDER BY r.created_at DESC`
	rows, err := mariadb.DB.Query(query, roomID)
	if err != nil {
		if err == sql.ErrNoRows {
			return records, nil
		}
		return records, err
	}
	defer rows.Close()
	for rows.Next() {
		var record todayRecord
		if err := rows.Scan(&record.RecordID, &record.Image); err != nil {
			return records, err
		}
		records = append(records, record)
	}
	return records, nil
}

func getDoneRecords(roomID int) ([]todayRecord, error) {
	records := []todayRecord{}
	query := `
		SELECT r.id, i.image
		FROM record r
		INNER JOIN ingredient i
		ON r.ingredient_id = i.id
		WHERE r.room_id = ? AND r.status = 1
		ORDER BY r.created_at DESC`
	rows, err := mariadb.DB.Query(query, roomID)
	if err != nil {
		if err == sql.ErrNoRows {
			return records, nil
		}
		return records, err
	}
	defer rows.Close()
	for rows.Next() {
		var record todayRecord
		if err := rows.Scan(&record.RecordID, &record.Image); err != nil {
			return records, err
		}
		records = append(records, record)
	}
	return records, nil
}

func JoinRoom(roomID int, userID int) error {
	// Check if room exists
	query := "SELECT EXISTS(SELECT 1 FROM room WHERE id = ?)"
	var exists bool
	err := mariadb.DB.QueryRow(query, roomID).Scan(&exists)
	if err != nil {
		return err
	} else if !exists {
		return fmt.Errorf("room does not exist")
	}
	// Check if user is already in room
	query = "SELECT EXISTS(SELECT 1 FROM room_user WHERE room_id = ? AND user_id = ?)"
	err = mariadb.DB.QueryRow(query, roomID, userID).Scan(&exists)
	if err != nil {
		return err
	} else if exists {
		return fmt.Errorf("user already in the room")
	}
	// Check if room is full
	query = "SELECT member_cnt, member_limit FROM room WHERE id = ?"
	var memberCnt, memberLimit int
	err = mariadb.DB.QueryRow(query, roomID).Scan(&memberCnt, &memberLimit)
	if err != nil {
		return err
	} else if memberCnt >= memberLimit {
		return fmt.Errorf("room is full")
	}
	// Begin transaction
	tx, err := mariadb.DB.Begin()
	if err != nil {
		return err
	}
	// Add user to room
	query = `
		INSERT INTO room_user (user_id, room_id)
		VALUES (?, ?)`
	_, err = mariadb.DB.Exec(query, userID, roomID)
	if err != nil {
		tx.Rollback()
		return err
	}
	// Update member count
	query = `
		UPDATE room
		SET member_cnt = member_cnt + 1
		WHERE id = ?`
	_, err = mariadb.DB.Exec(query, roomID)
	if err != nil {
		tx.Rollback()
		return err
	}
	// Commit transaction
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func LeaveRoom(roomID int, userID int) error {
	// Check if room exists
	query := "SELECT EXISTS(SELECT 1 FROM room WHERE id = ?)"
	var exists bool
	err := mariadb.DB.QueryRow(query, roomID).Scan(&exists)
	if err != nil {
		return err
	} else if !exists {
		return fmt.Errorf("room does not exist")
	}
	// Check if user is in room
	query = "SELECT EXISTS(SELECT 1 FROM room_user WHERE room_id = ? AND user_id = ?)"
	err = mariadb.DB.QueryRow(query, roomID, userID).Scan(&exists)
	if err != nil {
		return err
	} else if !exists {
		return fmt.Errorf("user not in room")
	}
	// Begin transaction
	tx, err := mariadb.DB.Begin()
	if err != nil {
		return err
	}
	// Remove user from room
	query = `
		DELETE FROM room_user
		WHERE user_id = ? AND room_id = ?`
	_, err = mariadb.DB.Exec(query, userID, roomID)
	if err != nil {
		tx.Rollback()
		return err
	}
	// Update member count
	query = `
		UPDATE room
		SET member_cnt = member_cnt - 1
		WHERE id = ?`
	_, err = mariadb.DB.Exec(query, roomID)
	if err != nil {
		tx.Rollback()
		return err
	}
	// Commit transaction
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}
	return nil
}
