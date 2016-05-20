package models

import (
	"github.com/jinzhu/gorm"
)

type User struct {
	gorm.Model
	Email    string
	Name     string
	Phone    string
	DLNumber string
	Cars     []Car `gorm:"ForeignKey:UserID"`
}

type Car struct {
	gorm.Model
	Brand              string
	CarModel           string
	Seats              uint8
	LicensePlateNumber string
	UserID             uint
}

type DbUser struct {
	User
	Hash []byte
	Salt []byte
}

func (u DbUser) TableName() string {
	return "users"
}

func DbUp(db *gorm.DB) {
	db.AutoMigrate(&DbUser{})
	db.AutoMigrate(&Car{})
	return
}
