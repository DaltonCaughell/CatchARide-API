package controllers

import (
	"CatchARide-API/models"
	"net/http"

	"github.com/jamieomatthews/validation"
	"github.com/jinzhu/gorm"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/render"
)

type AddCarData struct {
	DLNumber           string `form:"DLNumber" binding:"required"`
	Brand              string `form:"Brand" binding:"required"`
	Model              string `form:"Model" binding:"required"`
	Seats              uint8  `form:"Seats" binding:"required"`
	LicensePlateNumber string `form:"LicensePlateNumber" binding:"required"`
}

func (data *AddCarData) Validate(errors binding.Errors, req *http.Request) binding.Errors {

	v := validation.NewValidation(&errors, data)

	v.Validate(&data.DLNumber).Range(2, 255)
	v.Validate(&data.Brand).Range(2, 255)
	v.Validate(&data.Model).Range(2, 255)
	v.Validate(&data.LicensePlateNumber).Range(2, 255)
	if data.Seats < 1 {
		v.Errors.Add([]string{"Seats"}, "Validation Error", "Seat count cannot be less than 1")
	}

	return *v.Errors.(*binding.Errors)
}

func Me(r render.Render, user *models.DbUser) {
	r.JSON(200, user.User)
}

func AddCar(r render.Render, user *models.DbUser, data AddCarData, db *gorm.DB) {
	if data.DLNumber != "" {
		user.DLNumber = data.DLNumber
		db.Save(user)
	}
	car := &models.Car{Brand: data.Brand, CarModel: data.Model, LicensePlateNumber: data.LicensePlateNumber, Seats: data.Seats}
	db.Create(car)
	r.JSON(200, car)
}
