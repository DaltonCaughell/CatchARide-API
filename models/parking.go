package models

import (
	"math/rand"

	"time"

	"github.com/jinzhu/gorm"
)

type ParkingLot struct {
	gorm.Model
	Name      string
	Available uint16
	Total     uint16
	Lat       float64
	Lon       float64
	Open      uint8
	Close     uint8
	IsOpen    bool `gorm:"-"`
}

type ParkingLotNotification struct {
	gorm.Model
	UserID       uint
	ParkingLotID uint
}

func FakeParking(db *gorm.DB) {
	for {
		var lots []ParkingLot
		db.Find(&lots)
		for _, lot := range lots {
			lot.Available = uint16(rand.Intn(int(lot.Total)))
			db.Save(&lot)
		}
		time.Sleep(5 * time.Second)
	}
}
