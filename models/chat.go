package models

import "github.com/jinzhu/gorm"

type ChatMessage struct {
	gorm.Model
	ChatID     uint
	FromUserID uint
	Message    string
	ToUserID   uint
	User       User
}

type Chat struct {
	gorm.Model
}

type UserChat struct {
	gorm.Model
	ChatID uint
	UserID uint
}
