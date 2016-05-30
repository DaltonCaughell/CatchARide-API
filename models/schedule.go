package models

import (
	"CatchARide-API/config"
	"time"

	"github.com/jinzhu/gorm"
)

type ScheduledRide struct {
	gorm.Model
	UserID         uint
	CarID          uint
	From           string
	To             string
	DateTime       time.Time `gorm:"index"`
	FromLon        float64
	FromLat        float64
	ToLon          float64
	ToLat          float64
	ChatID         uint
	Seats          uint8
	Canceled       bool
	RatingsSent    bool
	Car            Car
	User           User
	DistFrom       float64 `gorm:"-"`
	Approved       bool    `gorm:"-"`
	Left           bool    `gorm:"-"`
	UnreadMessages bool    `gorm:"-"`
	Passengers     []Passenger
}

type RideSearch struct {
	gorm.Model
	UserID   uint
	From     string
	To       string
	DateTime time.Time `gorm:"index"`
	FromLon  float64
	FromLat  float64
	ToLon    float64
	ToLat    float64
	User     User
	Notify   bool
}

type Passenger struct {
	gorm.Model
	UserID       uint
	DriverID     uint
	RideID       uint
	Approved     bool
	RideSearchID uint
	Canceled     bool
	Details      RideSearch
	DistFrom     float64 `gorm:"-"`
	User         User
}

func SendRatings(db *gorm.DB) {
	for {
		currTime := time.Now().UTC()
		var rides []ScheduledRide
		db.Where("date_time < ? AND ratings_sent = ? AND canceled = ?", currTime.Format(config.MYSQL_DATE_FORMAT), false, false).Find(&rides)
		for i, ride := range rides {
			var passengers []Passenger
			db.Where("ride_id = ? AND approved = ? AND canceled = ?", ride.ID, true, false).Find(&passengers)
			db.Where("id = ?", ride.UserID).First(&ride.User)
			for _, p := range passengers {
				db.Where("id = ?", p.UserID).First(&p.User)
				message := &ChatMessage{
					ChatID:     ride.ChatID,
					Message:    "Rate Passenger " + p.User.Name,
					FromUserID: 0,
					ToUserID:   ride.UserID,
					Type:       "rating_req",
					LinkedID:   p.UserID,
				}
				db.Save(message)
				message = &ChatMessage{
					ChatID:     ride.ChatID,
					Message:    "Rate Driver " + ride.User.Name,
					FromUserID: 0,
					ToUserID:   p.UserID,
					Type:       "rating_req",
					LinkedID:   ride.UserID,
				}
				db.Save(message)
			}
			rides[i].RatingsSent = true
			db.Save(&rides[i])
		}
		time.Sleep(5 * time.Second)
	}
}
