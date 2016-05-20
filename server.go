package main

import (
	"CatchARide-API/controllers"
	"CatchARide-API/middleware"
	"CatchARide-API/models"
	"log"

	"github.com/go-martini/martini"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/cors"
	"github.com/martini-contrib/render"
	"gopkg.in/gcfg.v1"
)

type EnvConfig struct {
	DB struct {
		Username string
		Password string
		Address  string
		Port     string
		DbName   string
	}
}

type GlobalConfig struct {
	Test struct {
		Good bool
	}
}

var MAINTENANCE_MODE bool = false

func main() {
	m := martini.Classic()

	m.Use(render.Renderer())

	m.Use(cors.Allow(&cors.Options{
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{"GET", "POST"},
		AllowHeaders:  []string{"Origin", "Content-Type"},
		ExposeHeaders: []string{"Content-Length"},
	}))

	var globalConfig GlobalConfig

	err := gcfg.ReadFileInto(&globalConfig, "config/global.conf")

	if err != nil || !globalConfig.Test.Good {
		log.Printf("Global Config Load Fail: %s", err.Error())
		MAINTENANCE_MODE = true
	}

	var envConfig EnvConfig

	if martini.Env == "development" {
		err = gcfg.ReadFileInto(&envConfig, "config/dev.conf")
	} else {
		err = gcfg.ReadFileInto(&envConfig, "config/prod.conf")
	}

	if err != nil {
		log.Printf("Env Config Load Fail: %s", err.Error())
		MAINTENANCE_MODE = true
	}

	db, err := gorm.Open("mysql", envConfig.DB.Username+":"+envConfig.DB.Password+"@tcp("+envConfig.DB.Address+":"+envConfig.DB.Port+")/"+envConfig.DB.DbName+"?charset=utf8&parseTime=True")

	models.DbUp(&db)

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
				r.Post("/login", controllers.Login)
				r.Post("/create", binding.Bind(controllers.CreateData{}), controllers.Create)
			})
			r.Group("/*", func(r martini.Router) {
			}, middleware.BasicAuth)
		})
	})

	m.Get("/", func() string {
		return "Status: Good"
	})

	m.Run()
}
