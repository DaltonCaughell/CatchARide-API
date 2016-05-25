package controllers

import (
	"CatchARide-API/models"
	"net/http"

	"time"

	"errors"
	"log"

	"github.com/jamieomatthews/validation"
	"github.com/jinzhu/gorm"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/render"
	"golang.org/x/net/context"
	"googlemaps.github.io/maps"
)

type SearchData struct {
	IsDriver       bool   `form:"IsDriver" binding:"required"`
	From           string `form:"From" binding:"required"`
	To             string `form:"To" binding:"required"`
	DateTimeString string `form:"DateTime" binding:"required"`
	DateTime       time.Time
	FromLon        float64
	FromLat        float64
	ToLon          float64
	ToLat          float64
}

func geoCode(address string) (float64, float64, error) {
	c, err := maps.NewClient(maps.WithAPIKey("AIzaSyB8xMuFge6YmanK7_4kQFlFdkdvvDLhZSE"), maps.WithRateLimit(2))
	if err != nil {
		log.Fatalf("fatal error: %s", err)
		return 0, 0, err
	}
	r := &maps.GeocodingRequest{
		Address: address,
	}
	resp, err := c.Geocode(context.Background(), r)
	if err != nil {
		log.Fatalf("fatal error: %s", err)
		return 0, 0, err
	}
	if len(resp) == 0 {
		return 0, 0, errors.New("No Response")
	}

	return resp[0].Geometry.Location.Lat, resp[0].Geometry.Location.Lng, nil
}

func (data *SearchData) Validate(errors binding.Errors, req *http.Request) binding.Errors {

	var err error

	v := validation.NewValidation(&errors, data)

	if data.From != "SCHOOL" && data.To != "SCHOOL" {
		v.Errors.Add([]string{"From"}, "Validation Error", "At least one address must be school")
	}

	if data.From == "SCHOOL" {
		data.FromLat = globalConfig.Locations.SchoolLat
		data.FromLon = globalConfig.Locations.SchoolLat
	} else {
		if data.FromLat, data.FromLon, err = geoCode(data.From); err != nil {
			v.Errors.Add([]string{"From"}, "Validation Error", "Invalid from address")
		}
	}

	if data.To == "SCHOOL" {
		data.ToLat = globalConfig.Locations.SchoolLat
		data.ToLon = globalConfig.Locations.SchoolLat
	} else {
		if data.ToLat, data.ToLon, err = geoCode(data.To); err != nil {
			v.Errors.Add([]string{"To"}, "Validation Error", "Invalid to address")
		}
	}

	layout := "2006-01-02T15:04:05.000Z"
	data.DateTime, err = time.Parse(layout, data.DateTimeString)
	if err != nil {
		v.Errors.Add([]string{"Date"}, "Validation Error", "Invalid Date")
	}

	return *v.Errors.(*binding.Errors)
}

func Search(r render.Render, user *models.DbUser, db *gorm.DB, data SearchData) {

	if data.IsDriver {
		if len(user.Cars) == 0 {
			v := validation.NewValidation(new(binding.Errors), data)
			v.Errors.Add([]string{"Car"}, "Validation Error", "You do not have any registered cars")
			r.JSON(422, v.Errors)
			return
		}

		chat := &models.Chat{}
		db.Save(chat)

		userChat := &models.UserChat{
			UserID: user.ID,
			ChatID: chat.ID,
		}
		db.Save(userChat)

		ride := &models.ScheduledRide{
			UserID:   user.ID,
			CarID:    user.Cars[0].ID,
			From:     data.From,
			To:       data.To,
			DateTime: data.DateTime,
			FromLat:  data.FromLat,
			FromLon:  data.FromLon,
			ToLat:    data.ToLat,
			ToLon:    data.ToLon,
			ChatID:   chat.ID,
		}
		db.Save(ride)

		message := &models.ChatMessage{
			ChatID:     chat.ID,
			ToUserID:   user.ID,
			FromUserID: 0,
			Message:    "Your Ride Has Been Scheduled!",
		}
		db.Save(message)

		r.JSON(200, ride)
	} else {
		r.JSON(200, nil)
	}

}
