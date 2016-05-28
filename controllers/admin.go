package controllers

import (
	"CatchARide-API/models"

	"github.com/jinzhu/gorm"
	"github.com/martini-contrib/render"
)

func SeedRides(r render.Render, user *models.DbUser, db *gorm.DB) {
	r.JSON(200, struct{}{})
}
