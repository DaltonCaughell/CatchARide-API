package controllers

import (
	"CatchARide-API/models"
	"strconv"

	"fmt"

	"github.com/go-martini/martini"
	"github.com/jinzhu/gorm"
	"github.com/martini-contrib/render"
)

type SendData struct {
	Message string `form:"Message" binding:"required"`
}

type RequestCashData struct {
	Amount float64
}

func CashRequestReject(r render.Render, user *models.DbUser, db *gorm.DB, params martini.Params) {
	req := &models.ChatMessage{}
	db.Where("id = ?", params["MessageID"]).First(req)
	db.Where("id = ?", req.LinkedID).First(&req.CashRequest)
	req.CashRequest.Approved = false
	req.Ack = true
	message := &models.ChatMessage{
		ChatID:     req.ChatID,
		Message:    fmt.Sprintf("%s has rejected your request for $%.2f", user.Name, req.CashRequest.Amount),
		FromUserID: user.ID,
		ToUserID:   req.FromUserID,
		Type:       "",
		LinkedID:   req.CashRequest.ID,
	}
	db.Save(message)
	db.Save(req)
	db.Save(&req.CashRequest)
	r.JSON(200, struct{}{})
}

func CashRequestAccept(r render.Render, user *models.DbUser, db *gorm.DB, params martini.Params) {
	req := &models.ChatMessage{}
	toUser := &models.DbUser{}
	db.Where("id = ?", params["MessageID"]).First(req)
	db.Where("id = ?", req.LinkedID).First(&req.CashRequest)
	db.Where("id = ?", req.FromUserID).First(toUser)
	req.CashRequest.Approved = false
	req.Ack = true
	message := &models.ChatMessage{
		ChatID:     req.ChatID,
		Message:    fmt.Sprintf("%s has accepted your request for $%.2f", user.Name, req.CashRequest.Amount),
		FromUserID: user.ID,
		ToUserID:   req.FromUserID,
		Type:       "",
		LinkedID:   req.CashRequest.ID,
	}
	db.Save(message)
	message = &models.ChatMessage{
		ChatID:     req.ChatID,
		Message:    fmt.Sprintf("You sent %s $%.2f", toUser.Name, req.CashRequest.Amount),
		FromUserID: 0,
		ToUserID:   user.ID,
		LinkedID:   req.CashRequest.ID,
		Type:       "cash_info",
	}
	db.Save(message)
	user.Balance -= req.CashRequest.Amount
	toUser.Balance += req.CashRequest.Amount
	db.Save(req)
	db.Save(&req.CashRequest)
	db.Save(user)
	db.Save(toUser)
	r.JSON(200, struct{}{})
}

func RequestCash(r render.Render, user *models.DbUser, db *gorm.DB, params martini.Params, data RequestCashData) {
	ride := &models.ScheduledRide{}
	db.Where("chat_id = ?", params["ChatID"]).First(ride)
	var passengers []models.Passenger
	db.Where("ride_id = ? AND approved = ? AND canceled = ?", ride.ID, true, false).Find(&passengers)
	for _, p := range passengers {
		if p.UserID == user.ID {
			continue
		}
		request := &models.CashRequest{
			UserID:   user.ID,
			ToUserID: p.UserID,
			Amount:   data.Amount,
		}
		db.Save(request)
		message := &models.ChatMessage{
			ChatID:     ride.ChatID,
			Message:    fmt.Sprintf("%s has requested $%.2f from you", user.Name, request.Amount),
			FromUserID: user.ID,
			ToUserID:   p.UserID,
			Type:       "cash_req",
			LinkedID:   request.ID,
		}
		db.Save(message)
	}
	if ride.UserID != user.ID {
		request := &models.CashRequest{
			UserID:   user.ID,
			ToUserID: ride.UserID,
			Amount:   data.Amount,
		}
		db.Save(request)
		message := &models.ChatMessage{
			ChatID:     ride.ChatID,
			Message:    fmt.Sprintf("%s has requested $%.2f from you", user.Name, request.Amount),
			FromUserID: user.ID,
			ToUserID:   ride.UserID,
			Type:       "cash_req",
			LinkedID:   request.ID,
		}
		db.Save(message)
	}
	message := &models.ChatMessage{
		ChatID:     ride.ChatID,
		Message:    fmt.Sprintf("You requested $%.2f from the group", data.Amount),
		FromUserID: 0,
		ToUserID:   user.ID,
		LinkedID:   0,
		Type:       "cash_info",
	}
	db.Save(message)
}

func Rate(r render.Render, user *models.DbUser, db *gorm.DB, params martini.Params) {
	RatingID, _ := strconv.ParseInt(params["RatingID"], 10, 64)
	RatingValue, _ := strconv.ParseInt(params["Rating"], 10, 64)
	rating := &models.Rating{}
	if RatingID == 0 {
		message := &models.ChatMessage{}
		if !db.Where("id = ?", params["MessageID"]).First(message).RecordNotFound() {
			ride := &models.ScheduledRide{}
			if !db.Where("chat_id = ?", message.ChatID).First(ride).RecordNotFound() {
				rating = &models.Rating{
					UserID:      user.ID,
					RatedUserID: message.LinkedID,
					RideID:      ride.ID,
					Rating:      uint8(RatingValue),
				}
				db.Save(rating)
			}
		}
	} else {
		if !db.Where("id = ?", RatingID).First(rating).RecordNotFound() {
			rating.Rating = uint8(RatingValue)
			db.Save(rating)
		}
	}
	r.JSON(200, rating)
}

func Messages(r render.Render, user *models.DbUser, db *gorm.DB, params martini.Params) {
	var messages []models.ChatMessage
	ride := &models.ScheduledRide{}

	db.Model(&models.ScheduledRide{}).Where("chat_id = ?", params["ChatID"]).First(ride)
	db.Where("id = ?", ride.CarID).First(&ride.Car)
	db.Model(&models.DbUser{}).Where("id = ?", ride.UserID).First(&ride.User)

	if ride.UserID == user.ID {
		ride.Approved = true
		db.Where("chat_id = ? && (to_user_id = ? || to_user_id = ?)", params["ChatID"], user.ID, 0).Find(&messages)
	} else {
		passenger := &models.Passenger{}
		if !db.Where("ride_id = ? AND user_id = ? AND approved = ? AND canceled <> ?", ride.ID, user.ID, true, true).First(passenger).RecordNotFound() {
			ride.Approved = true
			db.Where("chat_id = ? && (to_user_id = ? || to_user_id = ?)", params["ChatID"], user.ID, 0).Find(&messages)
		} else {
			db.Where("chat_id = ? && (to_user_id = ?)", params["ChatID"], user.ID).Find(&messages)
		}
	}

	for index, message := range messages {
		db.Where("id = ?", message.FromUserID).First(&messages[index].User)
		if message.Type == "rating_req" {
			rating := &models.Rating{}
			if !db.Where("user_id = ? AND rated_user_id = ? AND ride_id = ?", user.ID, message.LinkedID, ride.ID).First(rating).RecordNotFound() {
				messages[index].Rating = *rating
			} else {
				messages[index].Rating = models.Rating{}
			}
		}
	}

	r.JSON(200, struct {
		Messages []models.ChatMessage
		Ride     *models.ScheduledRide
	}{
		messages, ride,
	})
}

func Send(r render.Render, user *models.DbUser, db *gorm.DB, params martini.Params, data SendData) {
	ChatID, _ := strconv.ParseInt(params["ChatID"], 10, 64)
	message := &models.ChatMessage{
		ChatID:     uint(ChatID),
		FromUserID: user.ID,
		ToUserID:   0,
		Message:    data.Message,
	}
	db.Save(message)
	r.JSON(200, message)
}
