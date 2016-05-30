package controllers

import (
	"CatchARide-API/lib/utils"
	"CatchARide-API/models"
	"fmt"
	"log"
	"net/http"
	"strings"

	"crypto/rand"
	"crypto/sha1"
	"io"

	"crypto/subtle"

	"github.com/jamieomatthews/validation"
	"github.com/jinzhu/gorm"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/render"
	"github.com/sendgrid/sendgrid-go"
	"golang.org/x/crypto/pbkdf2"
)

type SessionResponse struct {
	Response
	Session *models.Session
	User    *models.User
}

type LoginData struct {
	Email    string `form:"Email" binding:"required"`
	Password string `form:"Password" binding:"required"`
}

type CreateData struct {
	Name     string `form:"Name" binding:"required"`
	Email    string `form:"Email" binding:"required"`
	Password string `form:"Password" binding:"required"`
	Phone    string `form:"Phone" binding:"required"`
	Address  string `form:"Address" binding:"required"`
}

type ChangePasswordData struct {
	Password string `form:"Password" binding:"required"`
}

func (data *LoginData) Validate(errors binding.Errors, req *http.Request) binding.Errors {

	v := validation.NewValidation(&errors, data)

	data.Email = strings.ToLower(data.Email)

	v.Validate(&data.Email).TrimSpace().Email()
	v.Validate(&data.Password).Range(8, 255)

	return *v.Errors.(*binding.Errors)
}

func (data *CreateData) Validate(errors binding.Errors, req *http.Request) binding.Errors {

	v := validation.NewValidation(&errors, data)

	data.Email = strings.ToLower(data.Email)

	v.Validate(&data.Name).TrimSpace().Range(2, 255)
	v.Validate(&data.Email).TrimSpace().Email()
	v.Validate(&data.Password).Range(8, 255)
	v.Validate(&data.Phone).Range(10, 255)
	v.Validate(&data.Address).Range(10, 255)

	return *v.Errors.(*binding.Errors)
}

func (data *ChangePasswordData) Validate(errors binding.Errors, req *http.Request) binding.Errors {

	v := validation.NewValidation(&errors, data)

	v.Validate(&data.Password).Range(8, 255)

	return *v.Errors.(*binding.Errors)
}

type ResetData struct {
	Password string `form:"Password" binding:"required"`
	TempKey  string `form:"TempKey" binding:"required"`
}

func (data *ResetData) Validate(errors binding.Errors, req *http.Request) binding.Errors {

	v := validation.NewValidation(&errors, data)

	v.Validate(&data.Password).Range(8, 255)

	return *v.Errors.(*binding.Errors)
}

func Reset(r render.Render, data ResetData, db *gorm.DB) {
	forgot := &models.ForgotPassword{}
	if db.Where("temp_key = ?", data.TempKey).First(forgot).RecordNotFound() {
		r.JSON(200, struct{}{})
	} else {
		user := &models.DbUser{}
		db.Where("id = ?", forgot.UserID).First(user)
		salt := make([]byte, 32)
		_, err := io.ReadFull(rand.Reader, salt)
		if err != nil {
			r.JSON(500, Response{Code: 500, Error: "Internal Error", ErrorOn: ""})
			log.Print(err)
			return
		}
		hash := pbkdf2.Key([]byte(data.Password), salt, 4096, 255, sha1.New)

		user.Hash = hash
		user.Salt = salt

		db.Save(user)

		session, err := models.NewSession(db, user)
		if err != nil {
			r.JSON(500, Response{Code: 500, Error: "Internal Error", ErrorOn: ""})
			log.Print(err)
			return
		}

		forgot.Used = true

		db.Save(forgot)

		r.JSON(200, SessionResponse{Response: Response{Code: 0, Error: "", ErrorOn: ""}, Session: session, User: &user.User})
	}
}

type ForgotData struct {
	Email string `form:"Email" binding:"required"`
}

func (data *ForgotData) Validate(errors binding.Errors, req *http.Request) binding.Errors {

	v := validation.NewValidation(&errors, data)

	data.Email = strings.ToLower(data.Email)

	v.Validate(&data.Email).TrimSpace().Email()

	return *v.Errors.(*binding.Errors)
}

func Forgot(r render.Render, data ForgotData, db *gorm.DB, sg *sendgrid.SGClient) {
	user := &models.DbUser{}
	if !db.Where("email = ?", data.Email).First(user).RecordNotFound() {
		tempKey := utils.RandString(128)
		forgot := &models.ForgotPassword{}
		if db.Where("user_id = ? AND used = ?", user.ID, false).First(forgot).RecordNotFound() {
			forgot = &models.ForgotPassword{
				UserID: user.ID,
				Used:   false,
			}
		}
		forgot.TempKey = tempKey
		db.Save(forgot)
		message := sendgrid.NewMail()
		message.AddTo(user.Email)
		message.AddToName(user.Name)
		message.SetSubject("CatchARide Password Reset")
		message.SetFrom("admin@catcharide.today")
		message.SetHTML(fmt.Sprintf("<html><body>Click <a href='http://192.168.1.6:8000/#/reset/%s'>here</a> to set a new password.</body></html>", forgot.TempKey))
		if r := sg.Send(message); r == nil {
			fmt.Println("Email sent!")
		} else {
			fmt.Println(r)
		}
	}
	r.JSON(200, struct{}{})
}

func ChangePassword(r render.Render, user *models.DbUser, data ChangePasswordData, db *gorm.DB) {

	salt := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, salt)
	if err != nil {
		r.JSON(500, Response{Code: 500, Error: "Internal Error", ErrorOn: ""})
		log.Print(err)
		return
	}
	hash := pbkdf2.Key([]byte(data.Password), salt, 4096, 255, sha1.New)

	user.Hash = hash
	user.Salt = salt

	db.Save(user)

	r.JSON(200, user.User)
}

func Login(data LoginData, db *gorm.DB, r render.Render) {
	user := &models.DbUser{}
	if db.Model(&models.DbUser{}).Where("email = ?", data.Email).First(user).RecordNotFound() {
		v := validation.NewValidation(new(binding.Errors), data)
		v.Errors.Add([]string{"Email"}, "Validation Error", "Username/Password Incorrect")
		r.JSON(422, v.Errors)
		return
	}
	hash := pbkdf2.Key([]byte(data.Password), user.Salt, 4096, 255, sha1.New)
	if subtle.ConstantTimeCompare(hash, user.Hash) == 0 {
		v := validation.NewValidation(new(binding.Errors), data)
		v.Errors.Add([]string{"Email"}, "Validation Error", "Username/Password Incorrect")
		r.JSON(422, v.Errors)
		return
	} else {
		session, err := models.NewSession(db, user)
		if err != nil {
			r.JSON(500, Response{Code: 500, Error: "Internal Error", ErrorOn: ""})
			log.Print(err)
			return
		}
		r.JSON(200, SessionResponse{Response: Response{Code: 0, Error: "", ErrorOn: ""}, Session: session, User: &user.User})
	}
}

func Create(data CreateData, db *gorm.DB, r render.Render) {
	var count uint8
	db.Model(&models.DbUser{}).Where("email = ?", data.Email).Count(&count)
	if count != 0 {
		v := validation.NewValidation(new(binding.Errors), data)
		v.Errors.Add([]string{"Email"}, "Validation Error", "Email already exists")
		r.JSON(422, v.Errors)
		return
	} else {
		salt := make([]byte, 32)
		_, err := io.ReadFull(rand.Reader, salt)
		if err != nil {
			r.JSON(500, Response{Code: 500, Error: "Internal Error", ErrorOn: ""})
			log.Print(err)
			return
		}
		hash := pbkdf2.Key([]byte(data.Password), salt, 4096, 255, sha1.New)
		user := &models.DbUser{User: models.User{Name: data.Name, Email: data.Email, Phone: data.Phone, Address: data.Address}, Hash: hash, Salt: salt}
		db.Create(user)
		session, err := models.NewSession(db, user)
		if err != nil {
			r.JSON(500, Response{Code: 500, Error: "Internal Error", ErrorOn: ""})
			log.Print(err)
			return
		}
		r.JSON(200, SessionResponse{Response: Response{Code: 0, Error: "", ErrorOn: ""}, Session: session, User: &user.User})
	}
}
