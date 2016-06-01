package controllers

import (
	"CatchARide-API/models"
	"strconv"
	"time"

	"github.com/go-martini/martini"
	"github.com/jinzhu/gorm"
	"github.com/martini-contrib/render"
)

func SeedRides(r render.Render, db *gorm.DB, params martini.Params) {
	UserID, _ := strconv.ParseInt(params["UserID"], 10, 64)
	car := &models.Car{}
	db.Where("user_id = ?", UserID).First(&car)
	cTime := time.Now().UTC()
	for i := 0; i < 4*24*60; i += 20 {

		makeRide(db, "4337 8th AVE NE Seattle WA 98105", "SCHOOL", cTime, uint(UserID), car)
		makeRide(db, "SCHOOL", "4337 8th AVE NE Seattle WA 98105", cTime, uint(UserID), car)

		cTime = cTime.Add(time.Minute * 20)
	}
	r.JSON(200, struct{}{})
}

func makeRide(db *gorm.DB, from string, to string, cTime time.Time, UserID uint, car *models.Car) {

	chat := &models.Chat{}
	db.Save(chat)

	userChat := &models.UserChat{
		UserID: uint(UserID),
		ChatID: chat.ID,
	}
	db.Save(userChat)

	if from == "SCHOOL" {
		ride := &models.ScheduledRide{
			UserID:      uint(UserID),
			CarID:       car.ID,
			From:        from,
			To:          to,
			ToLat:       47.660595,
			ToLon:       -122.319803,
			FromLat:     47.657619,
			FromLon:     -122.307747,
			ChatID:      chat.ID,
			DateTime:    cTime,
			Seats:       4,
			Canceled:    false,
			RatingsSent: false,
		}
		db.Save(ride)
	} else {
		ride := &models.ScheduledRide{
			UserID:      uint(UserID),
			CarID:       car.ID,
			From:        from,
			To:          to,
			FromLat:     47.660595,
			FromLon:     -122.319803,
			ToLat:       47.657619,
			ToLon:       -122.307747,
			ChatID:      chat.ID,
			DateTime:    cTime,
			Seats:       4,
			Canceled:    false,
			RatingsSent: false,
		}
		db.Save(ride)
	}

	message := &models.ChatMessage{
		ChatID:     chat.ID,
		ToUserID:   uint(UserID),
		FromUserID: 0,
		Message:    "Your Ride Has Been Scheduled!",
	}
	db.Save(message)

	models.SetReadMessage(chat.ID, message.ID, uint(UserID), db)
}
