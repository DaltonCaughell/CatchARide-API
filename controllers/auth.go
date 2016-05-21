package controllers

import (
	"CatchARide-API/models"
	"log"
	"net/http"
	"strings"

	"crypto/rand"
	"crypto/sha1"
	"io"

	"crypto/subtle"

	"github.com/jamieomatthews/validation"
	"github.com/jinzhu/gorm"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/render"
	"golang.org/x/crypto/pbkdf2"
)

type SessionResponse struct {
	Response
	Session *models.Session
	User    *models.User
}

type LoginData struct {
	Email    string `form:"Email" binding:"required"`
	Password string `form:"Password" binding:"required"`
}

type CreateData struct {
	Name                  string `form:"Name" binding:"required"`
	Email                 string `form:"Email" binding:"required"`
	Password              string `form:"Password" binding:"required"`
	Phone                 string `form:"Phone" binding:"required"`
	Address               string `form:"Address" binding:"required"`
	DLNumber              string `form:"DLNumber"`
	CarBrand              string `form:"CarBrand"`
	CarModel              string `form:"CarModel"`
	CarSeats              uint8  `form:"CarSeats"`
	CarLicensePlateNumber string `form:"CarLicensePlateNumber"`
	IsDriver              bool   `form:"IsDriver"`
}

func (data *LoginData) Validate(errors binding.Errors, req *http.Request) binding.Errors {

	v := validation.NewValidation(&errors, data)

	data.Email = strings.ToLower(data.Email)

	v.Validate(&data.Email).TrimSpace().Email()
	v.Validate(&data.Password).Range(8, 255)

	return *v.Errors.(*binding.Errors)
}

func (data *CreateData) Validate(errors binding.Errors, req *http.Request) binding.Errors {

	v := validation.NewValidation(&errors, data)

	data.Email = strings.ToLower(data.Email)

	v.Validate(&data.Name).TrimSpace().Range(2, 255)
	v.Validate(&data.Email).TrimSpace().Email()
	v.Validate(&data.Password).Range(8, 255)
	v.Validate(&data.Phone).Range(10, 255)
	v.Validate(&data.Address).Range(10, 255)

	if data.IsDriver {
		v.Validate(&data.DLNumber).Range(2, 255)
		v.Validate(&data.CarBrand).Range(2, 255)
		v.Validate(&data.CarModel).Range(2, 255)
		v.Validate(&data.CarLicensePlateNumber).Range(2, 255)
		if data.CarSeats < 1 {
			v.Errors.Add([]string{"CarSeats"}, "Validation Error", "Seat count cannot be less than 1")
		}
	}

	return *v.Errors.(*binding.Errors)
}

func Login(data LoginData, db *gorm.DB, r render.Render) {
	user := &models.DbUser{}
	if db.Model(&models.DbUser{}).Where("email = ?", data.Email).First(user).RecordNotFound() {
		v := validation.NewValidation(new(binding.Errors), data)
		v.Errors.Add([]string{"Email"}, "Validation Error", "Username/Password Incorrect")
		r.JSON(422, v.Errors)
		return
	}
	hash := pbkdf2.Key([]byte(data.Password), user.Salt, 4096, 255, sha1.New)
	if subtle.ConstantTimeCompare(hash, user.Hash) == 0 {
		v := validation.NewValidation(new(binding.Errors), data)
		v.Errors.Add([]string{"Email"}, "Validation Error", "Username/Password Incorrect")
		r.JSON(422, v.Errors)
		return
	} else {
		session, err := models.NewSession(db, user)
		if err != nil {
			r.JSON(500, Response{Code: 500, Error: "Internal Error", ErrorOn: ""})
			log.Fatal(err)
			return
		}
		r.JSON(200, SessionResponse{Response: Response{Code: 0, Error: "", ErrorOn: ""}, Session: session, User: &user.User})
	}
}

func Create(data CreateData, db *gorm.DB, r render.Render) {
	var count uint8
	db.Model(&models.DbUser{}).Where("email = ?", data.Email).Count(&count)
	if count != 0 {
		v := validation.NewValidation(new(binding.Errors), data)
		v.Errors.Add([]string{"Email"}, "Validation Error", "Email already exists")
		r.JSON(422, v.Errors)
	} else {
		salt := make([]byte, 32)
		_, err := io.ReadFull(rand.Reader, salt)
		if err != nil {
			r.JSON(500, Response{Code: 500, Error: "Internal Error", ErrorOn: ""})
			log.Fatal(err)
			return
		}
		hash := pbkdf2.Key([]byte(data.Password), salt, 4096, 255, sha1.New)
		user := &models.DbUser{User: models.User{Name: data.Name, Email: data.Email, Phone: data.Phone, DLNumber: data.DLNumber}, Hash: hash, Salt: salt}
		db.Create(user)
		if data.IsDriver {
			db.Create(&models.Car{UserID: user.ID, Brand: data.CarBrand, CarModel: data.CarModel, Seats: data.CarSeats, LicensePlateNumber: data.CarLicensePlateNumber})
		}
		session, err := models.NewSession(db, user)
		if err != nil {
			r.JSON(500, Response{Code: 500, Error: "Internal Error", ErrorOn: ""})
			log.Fatal(err)
			return
		}
		r.JSON(200, SessionResponse{Response: Response{Code: 0, Error: "", ErrorOn: ""}, Session: session, User: &user.User})
	}
}
