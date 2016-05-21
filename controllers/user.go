package controllers

import (
	"CatchARide-API/models"

	"github.com/martini-contrib/render"
)

func Me(r render.Render, user *models.DbUser) {
	r.JSON(200, user.User)
}
