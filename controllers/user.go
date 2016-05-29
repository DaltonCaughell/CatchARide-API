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
	CarModel           string `form:"CarModel" binding:"required"`
	Seats              uint8  `form:"Seats" binding:"required"`
	LicensePlateNumber string `form:"LicensePlateNumber" binding:"required"`
}

func (data *AddCarData) Validate(errors binding.Errors, req *http.Request) binding.Errors {

	v := validation.NewValidation(&errors, data)

	v.Validate(&data.DLNumber).Range(2, 255)
	v.Validate(&data.Brand).Range(2, 255)
	v.Validate(&data.CarModel).Range(2, 255)
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
	car := &models.Car{Brand: data.Brand, CarModel: data.CarModel, LicensePlateNumber: data.LicensePlateNumber, Seats: data.Seats, UserID: user.ID}
	db.Create(car)
	r.JSON(200, car)
}

type UpdateData struct {
	Name     string `form:"Name" binding:"required"`
	Phone    string `form:"Phone" binding:"required"`
	Address  string `form:"Address" binding:"required"`
	DLNumber string `form:"DLNumber" binding:"required"`
}

func (data *UpdateData) Validate(errors binding.Errors, req *http.Request) binding.Errors {

	v := validation.NewValidation(&errors, data)

	v.Validate(&data.Name).TrimSpace().Range(2, 255)
	v.Validate(&data.Phone).Range(10, 255)
	v.Validate(&data.Address).Range(10, 255)
	v.Validate(&data.DLNumber).Range(2, 255)

	return *v.Errors.(*binding.Errors)
}

func UpdateUser(r render.Render, user *models.DbUser, data UpdateData, db *gorm.DB) {
	user.Address = data.Address
	user.Phone = data.Phone
	user.Name = data.Name
	user.DLNumber = data.DLNumber
	db.Save(user)
	r.JSON(200, user.User)
}

type UpdateCarData struct {
	Brand              string `form:"Brand" binding:"required"`
	CarModel           string `form:"CarModel" binding:"required"`
	Seats              uint8  `form:"Seats" binding:"required"`
	LicensePlateNumber string `form:"LicensePlateNumber" binding:"required"`
	ID                 uint8  `form:"ID"`
}

func (data *UpdateCarData) Validate(errors binding.Errors, req *http.Request) binding.Errors {

	v := validation.NewValidation(&errors, data)

	v.Validate(&data.Brand).Range(2, 255)
	v.Validate(&data.CarModel).Range(2, 255)
	v.Validate(&data.LicensePlateNumber).Range(2, 255)
	if data.Seats < 1 {
		v.Errors.Add([]string{"Seats"}, "Validation Error", "Seat count cannot be less than 1")
	}

	return *v.Errors.(*binding.Errors)
}

func UpdateCar(r render.Render, user *models.DbUser, data UpdateCarData, db *gorm.DB) {
	if data.ID == 0 && len(user.Cars) == 0 {
		user.Cars = append(user.Cars, models.Car{})
		user.Cars[0].UserID = user.ID
	}
	user.Cars[0].Brand = data.Brand
	user.Cars[0].CarModel = data.CarModel
	user.Cars[0].Seats = data.Seats
	user.Cars[0].LicensePlateNumber = data.LicensePlateNumber
	db.Save(&user.Cars[0])
	r.JSON(200, user.Cars[0])
}
