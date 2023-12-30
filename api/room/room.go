package room

import (
	"fmt"
	"pottogether/pkg/errhandler"
	"pottogether/pkg/logger"
	"pottogether/pkg/mariadb/query"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CreateRoomRequest struct {
	Name        string `json:"name"`
	MemberLimit int    `json:"memberLimit"`
	Privacy     string `json:"privacy"`
	Category    string `json:"category"`
}

func CreateRoom(c *gin.Context) {
	var req CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errhandler.Info(c, err, "Invalid request format")
		return
	}
	logger.Info("Request content: " + fmt.Sprintf("%+v", req))
	room := query.Room{
		ID:          -1,
		Name:        req.Name,
		MemberLimit: req.MemberLimit,
		Privacy:     req.Privacy,
		Category:    req.Category,
	}
	roomID, potID, err := query.CreateRoom(room, c.GetInt("id"))
	if err != nil {
		errhandler.Error(c, err, "Error creating room")
		return
	}
	c.JSON(200, gin.H{
		"isSuccess": true,
		"data": gin.H{
			"roomID": roomID,
			"potID":  potID,
		},
		"message": "Room created successfully",
	})
}

func GetRooms(c *gin.Context) {
	rooms, err := query.GetRooms(c.GetInt("id"))
	if err != nil {
		errhandler.Error(c, err, "Error getting rooms")
		return
	}
	c.JSON(200, gin.H{
		"isSuccess": true,
		"data":      rooms,
		"message":   "Rooms retrieved successfully",
	})
}

func GetPublicRooms(c *gin.Context) {
	rooms, err := query.GetPublicRooms()
	if err != nil {
		errhandler.Error(c, err, "Error getting public rooms")
		return
	}
	c.JSON(200, gin.H{
		"isSuccess": true,
		"data":      rooms,
		"message":   "Public rooms retrieved successfully",
	})
}

func GetRoomOverview(c *gin.Context) {
	roomID, err := strconv.Atoi(c.Param("roomID"))
	if err != nil {
		errhandler.Info(c, err, "Invalid roomID")
		return
	}
	room, err := query.GetRoomOverview(roomID, c.GetInt("id"))
	if err != nil {
		if err.Error() == "room does not exist" {
			errhandler.Info(c, err, "Room does not exist")
			return
		}
		errhandler.Error(c, err, "Error getting room overview")
		return
	}
	c.JSON(200, gin.H{
		"isSuccess": true,
		"data":      room,
		"message":   "Room overview retrieved successfully",
	})
}

func JoinRoom(c *gin.Context) {
	roomID, err := strconv.Atoi(c.Param("roomID"))
	if err != nil {
		errhandler.Info(c, err, "Invalid roomID")
		return
	}
	if err := query.JoinRoom(c.GetInt("id"), roomID); err != nil {
		if err.Error() == "room does not exist" {
			errhandler.Info(c, err, "Error joining room")
			return
		}
		if err.Error() == "user already in the room" {
			errhandler.Info(c, err, "Error joining room")
			return
		}
		if err.Error() == "room is full" {
			errhandler.Info(c, err, "Error joining room")
			return
		}
		errhandler.Error(c, err, "Error joining room")
		return
	}
}

func LeaveRoom(c *gin.Context) {
	roomID, err := strconv.Atoi(c.Param("roomID"))
	if err != nil {
		errhandler.Info(c, err, "Invalid roomID")
		return
	}
	if err := query.LeaveRoom(c.GetInt("id"), roomID); err != nil {
		if err.Error() == "room does not exist" {
			errhandler.Info(c, err, "Error leaving room")
			return
		}
		if err.Error() == "user not in room" {
			errhandler.Info(c, err, "Error leaving room")
			return
		}
		errhandler.Error(c, err, "Error leaving room")
		return
	}
}

func GetRoomRecords(c *gin.Context) {
	roomID, err := strconv.Atoi(c.Param("roomID"))
	if err != nil {
		errhandler.Info(c, err, "Invalid roomID")
		return
	}
	records, err := query.GetRoomRecords(roomID)
	if err != nil {
		errhandler.Error(c, err, "Error getting room records")
		return
	}
	c.JSON(200, gin.H{
		"isSuccess": true,
		"data":      records,
		"message":   "Room records retrieved successfully",
	})
}
