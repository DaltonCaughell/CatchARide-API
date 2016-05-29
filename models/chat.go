package models

import "github.com/jinzhu/gorm"

type ChatMessage struct {
	gorm.Model
	ChatID      uint
	FromUserID  uint
	Message     string
	ToUserID    uint
	User        User
	Type        string
	Style       string
	LinkedID    uint
	Rating      Rating
	CashRequest CashRequest
	Ack         bool
}

type Chat struct {
	gorm.Model
}

type UserChat struct {
	gorm.Model
	ChatID uint
	UserID uint
}

type CashRequest struct {
	gorm.Model
	Amount   float64
	UserID   uint
	ToUserID uint
	Approved bool
}
