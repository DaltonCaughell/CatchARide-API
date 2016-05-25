package controllers

import (
	"CatchARide-API/models"

	"github.com/go-martini/martini"
	"github.com/jinzhu/gorm"
	"github.com/martini-contrib/render"
)

func Messages(r render.Render, user *models.DbUser, db *gorm.DB, params martini.Params) {
	var messages []models.ChatMessage
	db.Where("chat_id = ? && (to_user_id = ? || to_user_id = ?)", params["ChatID"], user.ID, 0).Find(&messages)
	r.JSON(200, messages)
}
