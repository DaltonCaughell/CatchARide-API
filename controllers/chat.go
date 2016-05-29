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

func Rate(r render.Render, user *models.DbUser, db *gorm.DB, params martini.Params) {
	RatingID, _ := strconv.ParseInt(params["RatingID"], 10, 64)
	RatingValue, _ := strconv.ParseInt(params["Rating"], 10, 64)
	rating := &models.Rating{}
	if RatingID == 0 {
		message := &models.ChatMessage{}
		if !db.Where("id = ?", params["MessageID"]).First(message).RecordNotFound() {
			ride := &models.ScheduledRide{}
			if !db.Where("chat_id = ?", message.ChatID).First(ride).RecordNotFound() {
				rating = &models.Rating{
					UserID:      user.ID,
					RatedUserID: message.LinkedID,
					RideID:      ride.ID,
					Rating:      uint8(RatingValue),
				}
				db.Save(rating)
			}
		}
	} else {
		if !db.Where("id = ?", RatingID).First(rating).RecordNotFound() {
			rating.Rating = uint8(RatingValue)
			db.Save(rating)
		}
	}
	r.JSON(200, rating)
}

func Messages(r render.Render, user *models.DbUser, db *gorm.DB, params martini.Params) {
	var messages []models.ChatMessage
	ride := &models.ScheduledRide{}

	db.Model(&models.ScheduledRide{}).Where("chat_id = ?", params["ChatID"]).First(ride)
	db.Where("id = ?", ride.CarID).First(&ride.Car)
	db.Model(&models.DbUser{}).Where("id = ?", ride.UserID).First(&ride.User)

	if ride.UserID == user.ID {
		ride.Approved = true
		db.Where("chat_id = ? && (to_user_id = ? || to_user_id = ?)", params["ChatID"], user.ID, 0).Find(&messages)
	} else {
		passenger := &models.Passenger{}
		if !db.Where("ride_id = ? AND user_id = ? AND approved = ? AND canceled <> ?", ride.ID, user.ID, true, true).First(passenger).RecordNotFound() {
			ride.Approved = true
			db.Where("chat_id = ? && (to_user_id = ? || to_user_id = ?)", params["ChatID"], user.ID, 0).Find(&messages)
		} else {
			db.Where("chat_id = ? && (to_user_id = ?)", params["ChatID"], user.ID).Find(&messages)
		}
	}

	for index, message := range messages {
		db.Where("id = ?", message.FromUserID).First(&messages[index].User)
		if message.Type == "rating_req" {
			rating := &models.Rating{}
			if !db.Where("user_id = ? AND rated_user_id = ? AND ride_id = ?", user.ID, message.LinkedID, ride.ID).First(rating).RecordNotFound() {
				messages[index].Rating = *rating
			} else {
				messages[index].Rating = models.Rating{}
			}
		}
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
