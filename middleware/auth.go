package middleware

import (
	"CatchARide-API/controllers"
	"CatchARide-API/models"
	"net/http"

	"github.com/go-martini/martini"
	"github.com/jinzhu/gorm"
	"github.com/martini-contrib/render"
)

func BasicAuth(c martini.Context, req *http.Request, r render.Render, db *gorm.DB) {
	inKey := req.Header.Get("X-API-KEY")
	session := &models.Session{}
	if db.Model(&models.Session{}).Where("api_key = ?", inKey).First(session).RecordNotFound() {
		r.JSON(302, controllers.Response{Code: 0, Error: "Not Authorized", ErrorOn: ""})
	} else {
		user := &models.DbUser{}
		if db.Model(&models.DbUser{}).Where(&models.DbUser{User: models.User{Model: gorm.Model{ID: session.UserID}}}).First(user).Related(&user.Cars, "user_id").RecordNotFound() {
			r.JSON(302, controllers.Response{Code: 0, Error: "Not Authorized", ErrorOn: ""})
		} else {
			c.Map(user)
		}
	}
}
