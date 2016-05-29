package main

import (
	"CatchARide-API/controllers"
	"CatchARide-API/middleware"
	"CatchARide-API/models"
	"log"

	"CatchARide-API/config"

	"github.com/go-martini/martini"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/cors"
	"github.com/martini-contrib/render"
	"gopkg.in/gcfg.v1"
)

var MAINTENANCE_MODE bool = false

func main() {
	m := martini.Classic()

	m.Use(render.Renderer())

	m.Use(cors.Allow(&cors.Options{
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{"GET", "POST", "PUT"},
		AllowHeaders:  []string{"Origin", "Content-Type", "X-API-KEY"},
		ExposeHeaders: []string{"Content-Length"},
	}))

	var globalConfig config.GlobalConfig

	err := gcfg.ReadFileInto(&globalConfig, "config/global.conf")

	if err != nil || !globalConfig.Test.Good {
		log.Printf("Global Config Load Fail: %s", err.Error())
		MAINTENANCE_MODE = true
	}

	var envConfig config.EnvConfig

	if martini.Env == "development" {
		err = gcfg.ReadFileInto(&envConfig, "config/dev.conf")
	} else {
		err = gcfg.ReadFileInto(&envConfig, "config/prod.conf")
	}

	if err != nil {
		log.Printf("Env Config Load Fail: %s", err.Error())
		MAINTENANCE_MODE = true
	}

	controllers.RegConfig(globalConfig, envConfig)

	db, err := gorm.Open("mysql", envConfig.DB.Username+":"+envConfig.DB.Password+"@tcp("+envConfig.DB.Address+":"+envConfig.DB.Port+")/"+envConfig.DB.DbName+"?charset=utf8&parseTime=True")

	models.DbUp(db)

	if err != nil {
		log.Printf("DB Connection Error: %s", err.Error())
		MAINTENANCE_MODE = true
	}

	err = db.DB().Ping()
	if err != nil {
		log.Printf("DB Connection Error: %s", err.Error())
		MAINTENANCE_MODE = true
	}

	m.Use(func(r render.Render) {
		if MAINTENANCE_MODE {
			r.JSON(200, controllers.Response{Code: 501, Error: "Down For Maintenance", ErrorOn: ""})
		}
	})

	m.Map(db)

	m.Group("/api", func(r martini.Router) {
		r.Group("/v1", func(r martini.Router) {
			r.Group("/auth", func(r martini.Router) {
				r.Post("/login", binding.Bind(controllers.LoginData{}), controllers.Login)
				r.Post("/create", binding.Bind(controllers.CreateData{}), controllers.Create)
				r.Post("/password", binding.Bind(controllers.ChangePasswordData{}), middleware.BasicAuth, controllers.ChangePassword)
			})
			r.Group("/user", func(r martini.Router) {
				r.Get("/me", controllers.Me)
				r.Post("/addcar", binding.Bind(controllers.AddCarData{}), controllers.AddCar)
				r.Post("/me", binding.Bind(controllers.UpdateData{}), controllers.UpdateUser)
				r.Post("/car", binding.Bind(controllers.UpdateCarData{}), controllers.UpdateCar)
			}, middleware.BasicAuth)
			r.Group("/parking", func(r martini.Router) {
				r.Get("/all", controllers.All)
			}, middleware.BasicAuth)
			r.Group("/schedule", func(r martini.Router) {
				r.Post("/search", binding.Bind(controllers.SearchData{}), controllers.Search)
				r.Get("/me", controllers.GetScheduledRides)
				r.Get("/ride/:RideID", controllers.Ride)
				r.Get("/available/:SearchID", controllers.Available)
				r.Get("/join/:RideID/:SearchID", controllers.Join)
				r.Get("/leave/:RideID", controllers.Leave)
				r.Get("/acceptpassenger/:RideID/:MessageID", controllers.AcceptPassenger)
				r.Get("/rejectpassenger/:RideID/:MessageID", controllers.RejectPassenger)
				r.Get("/rider/:RideID/:UserID", controllers.Rider)
			}, middleware.BasicAuth)
			r.Group("/chat", func(r martini.Router) {
				r.Get("/messages/:ChatID", controllers.Messages)
				r.Post("/send/:ChatID", binding.Bind(controllers.SendData{}), controllers.Send)
				r.Put("/rate/:MessageID/:RatingID/:Rating", controllers.Rate)
				r.Post("/requestcash/:ChatID", binding.Bind(controllers.RequestCashData{}), controllers.RequestCash)
				r.Put("/cashrequestaccept/:MessageID", controllers.CashRequestAccept)
				r.Put("/cashrequestreject/:MessageID", controllers.CashRequestReject)
			}, middleware.BasicAuth)
			r.Group("/*", func(r martini.Router) {
			}, middleware.BasicAuth)
		})
	})

	m.Get("/", func() string {
		return "Status: Good"
	})

	go models.FakeParking(db)
	go models.SendRatings(db)

	m.Run()
}
