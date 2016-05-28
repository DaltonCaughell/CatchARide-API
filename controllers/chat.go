package controllers

import (
	"CatchARide-API/models"
	"strconv"

	"github.com/go-martini/martini"
	"github.com/jinzhu/gorm"
	"github.com/martini-contrib/render"
)

type SendData struct {
	Message string `form:"Message" binding:"required"`
}

func Messages(r render.Render, user *models.DbUser, db *gorm.DB, params martini.Params) {
	var messages []models.ChatMessage
	ride := &models.ScheduledRide{}

	db.Model(&models.ScheduledRide{}).Where("chat_id = ?", params["ChatID"]).First(ride)
	db.Where("id = ?", ride.CarID).First(&ride.Car)
	db.Model(&models.DbUser{}).Where("id = ?", ride.UserID).First(&ride.User)

	if ride.UserID == user.ID {
		db.Where("chat_id = ? && (to_user_id = ? || to_user_id = ?)", params["ChatID"], user.ID, 0).Find(&messages)
	} else {
		passenger := &models.Passenger{}
		if !db.Where("ride_id = ?", ride.ID).First(passenger).RecordNotFound() {
			if passenger.Approved {
				db.Where("chat_id = ? && (to_user_id = ? || to_user_id = ?)", params["ChatID"], user.ID, 0).Find(&messages)
			} else {
				db.Where("chat_id = ? && (to_user_id = ?)", params["ChatID"], user.ID).Find(&messages)
			}
		}
	}

	for index, message := range messages {
		db.Where("id = ?", message.FromUserID).First(&messages[index].User)
	}

	r.JSON(200, struct {
		Messages []models.ChatMessage
		Ride     *models.ScheduledRide
	}{
		messages, ride,
	})
}

func Send(r render.Render, user *models.DbUser, db *gorm.DB, params martini.Params, data SendData) {
	ChatID, _ := strconv.ParseInt(params["ChatID"], 10, 64)
	message := &models.ChatMessage{
		ChatID:     uint(ChatID),
		FromUserID: user.ID,
		ToUserID:   0,
		Message:    data.Message,
	}
	db.Save(message)
	r.JSON(200, message)
}
