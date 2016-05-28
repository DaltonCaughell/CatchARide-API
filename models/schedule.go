package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type ScheduledRide struct {
	gorm.Model
	UserID      uint
	CarID       uint
	From        string
	To          string
	DateTime    time.Time `gorm:"index"`
	FromLon     float64
	FromLat     float64
	ToLon       float64
	ToLat       float64
	ChatID      uint
	Seats       uint8
	Canceled    bool
	RatingsSent bool
	Car         Car
	User        User
	DistFrom    float64 `gorm:"-"`
	Approved    bool    `gorm:"-"`
	Passengers  []Passenger
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
	Details      RideSearch
	DistFrom     float64 `gorm:"-"`
	User         User
}
