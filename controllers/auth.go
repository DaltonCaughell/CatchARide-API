package controllers

import (
	_ "CatchARide-API/lib/validation"
	_ "CatchARide-API/models"
)

type Response struct {
	Code    int64
	Error   string
	ErrorOn string
}

type CreateData struct {
	NameFirst string `form:"nameFirst"`
	NameLast  string `form:"nameLast"`
	Email     string `form:"email"`
	Password  string `form:"password"`
	Phone     string `form:"phone"`
}

/*func (c CreateData) Validate(errors binding.Errors, req *http.Request) binding.Errors {

	var valid bool

	if !validation.ValidEmail(&c.Email, "Email", "email", &errors) {
		return errors
	} else if !validation.ValidName(&c.NameFirst, "First Name", "nameFirst", &errors) {
		return errors
	} else if !validation.ValidName(&c.NameLast, "Last Name", "nameLast", &errors) {
		return errors
	} else if !validation.ValidPassword(&c.Password, "Password", "password", &errors) {
		return errors
	} else if c.Phone, valid = validation.ValidPhone(c.Phone, "Phone Number", "phone", &errors); !valid {
		return errors
	}

	return errors
}*/

func Login() {

}

func Create(data CreateData) {

}
