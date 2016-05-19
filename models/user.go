package models

import (
	"github.com/jinzhu/gorm"
)

type User struct {
	gorm.Model
	Email string
	Name  string
	Phone string
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
	return
}
