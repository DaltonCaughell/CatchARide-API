package controllers

import (
	"CatchARide-API/config"
	"CatchARide-API/models"
	"net/http"

	"time"

	"errors"
	"log"

	"github.com/go-martini/martini"
	"github.com/jamieomatthews/validation"
	"github.com/jinzhu/gorm"
	"github.com/kellydunn/golang-geo"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/render"
	"golang.org/x/net/context"
	"googlemaps.github.io/maps"
)

type SearchData struct {
	IsDriver bool      `form:"IsDriver"`
	From     string    `form:"From" binding:"required"`
	To       string    `form:"To" binding:"required"`
	DateTime time.Time `form:"DateTime" binding:"required"`
	FromLon  float64
	FromLat  float64
	ToLon    float64
	ToLat    float64
}

func geoCode(address string) (float64, float64, error) {
	c, err := maps.NewClient(maps.WithAPIKey("AIzaSyB8xMuFge6YmanK7_4kQFlFdkdvvDLhZSE"), maps.WithRateLimit(2))
	if err != nil {
		log.Printf("fatal error: %s", err)
		return 0, 0, err
	}
	r := &maps.GeocodingRequest{
		Address: address,
	}
	resp, err := c.Geocode(context.Background(), r)
	if err != nil {
		log.Printf("fatal error geo-codeing: %s Address: %s", err, address)
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

	if len(errors) != 0 {
		return *v.Errors.(*binding.Errors)
	}

	if data.From != "SCHOOL" && data.To != "SCHOOL" {
		v.Errors.Add([]string{"From"}, "Validation Error", "At least one address must be school")
		return *v.Errors.(*binding.Errors)
	}

	if data.From == "SCHOOL" {
		data.FromLat = globalConfig.Locations.SchoolLat
		data.FromLon = globalConfig.Locations.SchoolLon
	} else {
		if data.FromLat, data.FromLon, err = geoCode(data.From); err != nil {
			v.Errors.Add([]string{"From"}, "Validation Error", "Invalid from address")
		}
	}

	if data.To == "SCHOOL" {
		data.ToLat = globalConfig.Locations.SchoolLat
		data.ToLon = globalConfig.Locations.SchoolLon
	} else {
		if data.ToLat, data.ToLon, err = geoCode(data.To); err != nil {
			v.Errors.Add([]string{"To"}, "Validation Error", "Invalid to address")
		}
	}

	log.Printf("%d", data.DateTime.Year())

	return *v.Errors.(*binding.Errors)
}

func Join(r render.Render, user *models.DbUser, db *gorm.DB, params martini.Params) {

	ride := &models.ScheduledRide{}

	db.Where("id = ?", params["RideID"]).First(ride)
	db.Where("id = ?", ride.CarID).First(&ride.Car)
	db.Model(&models.DbUser{}).Where("id = ?", ride.UserID).First(&ride.User)

	if ride.Seats <= 0 {
		r.JSON(422, struct{}{})
		return
	}

	driverMessage := &models.ChatMessage{
		ToUserID:   ride.User.ID,
		Message:    user.Name + " has requested a ride!",
		Type:       "ride_request",
		FromUserID: user.ID,
		ChatID:     ride.ChatID,
	}

	db.Save(driverMessage)

	riderMessage := &models.ChatMessage{
		ToUserID:   user.ID,
		Message:    "Pending approval to join " + ride.User.Name + "'s car...",
		Type:       "",
		FromUserID: 0,
		ChatID:     ride.ChatID,
	}

	db.Save(riderMessage)

	search := &models.RideSearch{}

	db.Where("id = ?", params["SearchID"]).First(search)

	passenger := &models.Passenger{
		UserID:       user.ID,
		RideID:       ride.ID,
		DriverID:     ride.User.ID,
		Approved:     false,
		RideSearchID: search.ID,
	}

	db.Save(passenger)

	ride.Seats--

	db.Save(ride)

	r.JSON(200, struct {
		Ride      *models.ScheduledRide
		Passenger *models.Passenger
	}{
		Ride:      ride,
		Passenger: passenger,
	})
}

func managePassenger(r render.Render, user *models.DbUser, db *gorm.DB, params martini.Params, ack bool) {
	message := &models.ChatMessage{}
	db.Where("id = ?", params["MessageID"]).First(message)
	if message.Ack {
		r.JSON(200, struct{}{})
		return
	}
	message.Ack = true
	db.Save(message)
	ride := &models.ScheduledRide{}
	db.Where("id = ?", params["RideID"]).First(&ride)
	db.Where("id = ?", ride.UserID).First(&ride.User)
	passenger := &models.DbUser{}
	db.Where("id = ?", message.FromUserID).First(passenger)
	var mText string
	if ack {
		mText = passenger.Name + " has joined your car!"
	} else {
		mText = passenger.Name + " was removed from your car!"
	}
	ackMessage := &models.ChatMessage{
		FromUserID: 0,
		ToUserID:   user.ID,
		ChatID:     message.ChatID,
		Message:    mText,
	}
	db.Save(ackMessage)
	if ack {
		mText = "Your ride request was approved!"
	} else {
		mText = "Your ride request was rejected!"
	}
	ackMessage = &models.ChatMessage{
		FromUserID: 0,
		ToUserID:   message.FromUserID,
		ChatID:     message.ChatID,
		Message:    mText,
	}
	db.Save(ackMessage)
	pRide := &models.Passenger{}
	db.Where("user_id = ? AND driver_id = ? && ride_id = ?", passenger.ID, user.ID, ride.ID).First(pRide)
	pRide.Approved = ack
	db.Save(pRide)
	if !ack {
		ride.Seats++
	}
	db.Save(ride)
	r.JSON(200, struct{}{})
}

func AcceptPassenger(r render.Render, user *models.DbUser, db *gorm.DB, params martini.Params) {
	managePassenger(r, user, db, params, true)
}

func RejectPassenger(r render.Render, user *models.DbUser, db *gorm.DB, params martini.Params) {
	managePassenger(r, user, db, params, false)
}

func Available(r render.Render, user *models.DbUser, db *gorm.DB, params martini.Params) {
	var rides []models.ScheduledRide
	data := &models.RideSearch{}
	db.Where("id = ?", params["SearchID"]).First(data)
	if data.From == "SCHOOL" {
		db.Where("`from` = ? AND date_time <= ? AND date_time > ? AND user_id  <> ? AND seats > ?", "SCHOOL",
			data.DateTime.Format(config.MYSQL_DATE_FORMAT), data.DateTime.Add(time.Minute*-30).Format(config.MYSQL_DATE_FORMAT), user.ID, 0).Find(&rides)
		for index, ride := range rides {
			p := geo.NewPoint(data.ToLat, data.ToLon)
			rides[index].DistFrom = p.GreatCircleDistance(geo.NewPoint(ride.ToLat, ride.ToLon))
		}
	} else {
		db.Where("`to` = ? AND date_time <= ? AND date_time > ? AND user_id  <> ? AND seats > ?", "SCHOOL",
			data.DateTime.Format(config.MYSQL_DATE_FORMAT), data.DateTime.Add(time.Minute*-30).Format(config.MYSQL_DATE_FORMAT), user.ID, 0).Find(&rides)
		for index, ride := range rides {
			p := geo.NewPoint(data.FromLat, data.FromLon)
			rides[index].DistFrom = p.GreatCircleDistance(geo.NewPoint(ride.FromLat, ride.FromLon))
		}
	}
	for index, ride := range rides {
		db.Where("id = ?", ride.CarID).First(&rides[index].Car)
		db.Model(&models.DbUser{}).Where("id = ?", ride.UserID).First(&rides[index].User)
	}
	r.JSON(200, struct {
		Rides  []models.ScheduledRide
		Search *models.RideSearch
	}{
		Rides:  rides,
		Search: data,
	})
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
			Seats:    user.Cars[0].Seats,
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
		search := &models.RideSearch{
			UserID:   user.ID,
			From:     data.From,
			To:       data.To,
			DateTime: data.DateTime,
			FromLat:  data.FromLat,
			FromLon:  data.FromLon,
			ToLat:    data.ToLat,
			ToLon:    data.ToLon,
			Notify:   false,
		}
		db.Save(search)
		r.JSON(200, search)
	}
}

func GetScheduledRides(r render.Render, user *models.DbUser, db *gorm.DB) {
	var rides []models.ScheduledRide
	db.Where("user_id = ?", user.ID).Find(&rides)
	var pRides []models.Passenger
	db.Where("user_id = ?", user.ID).Find(&pRides)
	for _, p := range pRides {
		ride := models.ScheduledRide{}
		db.Where("id = ?", p.RideID).First(&ride)
		ride.Approved = p.Approved
		rides = append(rides, ride)
	}
	for index, ride := range rides {
		db.Where("id = ?", ride.CarID).First(&rides[index].Car)
		db.Model(&models.DbUser{}).Where("id = ?", ride.UserID).First(&rides[index].User)
	}
	r.JSON(200, rides)
}

func Ride(r render.Render, user *models.DbUser, db *gorm.DB, params martini.Params) {
	ride := &models.ScheduledRide{}
	db.Where("id = ?", params["RideID"]).First(&ride)
	db.Where("id = ?", ride.CarID).First(&ride.Car)
	db.Model(&models.DbUser{}).Where("id = ?", ride.UserID).First(&ride.User)
	db.Where("ride_id = ? AND approved = ?", ride.ID, true).Find(&ride.Passengers)
	for index, p := range ride.Passengers {
		db.Where("id = ?", p.RideSearchID).First(&ride.Passengers[index].Details)
		if ride.From == "SCHOOL" {
			p := geo.NewPoint(ride.Passengers[index].Details.ToLat, ride.Passengers[index].Details.ToLon)
			ride.Passengers[index].DistFrom = p.GreatCircleDistance(geo.NewPoint(ride.ToLat, ride.ToLon))
		} else {
			p := geo.NewPoint(ride.Passengers[index].Details.FromLat, ride.Passengers[index].Details.FromLon)
			ride.Passengers[index].DistFrom = p.GreatCircleDistance(geo.NewPoint(ride.FromLat, ride.FromLon))
		}
		db.Model(&models.DbUser{}).Where("id = ?", p.UserID).First(&ride.Passengers[index].User)
	}
	r.JSON(200, ride)
}
