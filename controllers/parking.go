package controllers

import (
	"CatchARide-API/models"

	"github.com/jinzhu/gorm"
	"github.com/martini-contrib/render"
)

func All(r render.Render, user *models.DbUser, db *gorm.DB) {
	var lots []models.ParkingLot
	db.Find(&lots)
	r.JSON(200, lots)
}
