package controllers

import (
	"CatchARide-API/models"

	"github.com/go-martini/martini"
	"github.com/jinzhu/gorm"
	"github.com/martini-contrib/render"
)

func All(r render.Render, user *models.DbUser, db *gorm.DB) {
	var lots []models.ParkingLot
	db.Find(&lots)
	for index := range lots {
		if db.Where("user_id = ? AND parking_lot_id = ?", user.ID, lots[index].ID).First(&lots[index].Notification).RecordNotFound() {
			lots[index].Notification = models.ParkingLotNotification{
				UserID:       user.ID,
				ParkingLotID: lots[index].ID,
				Notify:       false,
			}
			db.Save(&lots[index].Notification)
		}
	}
	r.JSON(200, lots)
}

func SetParkingNotify(r render.Render, user *models.DbUser, db *gorm.DB, params martini.Params) {
	notify := &models.ParkingLotNotification{}
	db.Where("user_id = ? AND parking_lot_id = ?", user.ID, params["LotID"]).First(notify)
	if params["Notify"] == "true" {
		notify.Notify = true
	} else {
		notify.Notify = false
	}
	db.Save(notify)
	r.JSON(200, notify)
}
