package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type ScheduledRide struct {
	gorm.Model
	UserID   uint
	CarID    uint
	From     string
	To       string
	DateTime time.Time
	FromLon  float64
	FromLat  float64
	ToLon    float64
	ToLat    float64
	ChatID   uint
	Car      Car
	User     User
	DistFrom float64 `gorm:"-"`
}
