package models

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"log"

	"github.com/jinzhu/gorm"
)

type User struct {
	gorm.Model
	Email    string
	Name     string
	Phone    string
	DLNumber string
	Cars     []Car     `gorm:"ForeignKey:UserID"`
	Sessions []Session `gorm:"ForeignKey:UserID"`
}

type Car struct {
	gorm.Model
	Brand              string
	CarModel           string
	Seats              uint8
	LicensePlateNumber string
	UserID             uint
}

type Session struct {
	gorm.Model
	UserID uint
	Key    string
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
	db.AutoMigrate(&Session{})
	return
}

func NewSession(db *gorm.DB, user *DbUser) (*Session, error) {
	base := make([]byte, 128)
	_, err := io.ReadFull(rand.Reader, base)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	session := &Session{UserID: user.User.ID, Key: base64.RawStdEncoding.EncodeToString(base)}
	db.Create(session)
	return session, nil
}
