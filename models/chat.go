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

type ReadMessage struct {
	gorm.Model
	ChatID    uint
	UserID    uint
	MessageID uint
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

func HasUnreadMessages(ChatID uint, UserID uint, All bool, db *gorm.DB) bool {
	unread := false
	var messages []ChatMessage
	if All {
		db.Where("chat_id = ? AND (to_user_id = ? OR to_user_id = ?)", ChatID, UserID, 0).Find(&messages)
	} else {
		db.Where("chat_id = ? AND (to_user_id = ?)", ChatID, UserID).Find(&messages)
	}
	for _, m := range messages {
		read := &ReadMessage{}
		if db.Where("chat_id = ? AND user_id = ? AND message_id = ?", m.ChatID, UserID, m.ID).First(read).RecordNotFound() {
			unread = true
			break
		}
	}
	return unread
}

func SetReadMessage(ChatID uint, MessageID uint, UserID uint, db *gorm.DB) *ReadMessage {
	read := &ReadMessage{}
	db.FirstOrCreate(read, ReadMessage{MessageID: MessageID, UserID: UserID, ChatID: ChatID})
	return read
}
