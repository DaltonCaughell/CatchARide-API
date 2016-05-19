package validation

import (
	"errors"
	"regexp"
	"strconv"
	_ "strings"
)

var emailRegexString string = `^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+$`

var EmailRegexp *regexp.Regexp

func init() {
	var err error
	EmailRegexp, err = regexp.Compile(emailRegexString)
	if err != nil {
		panic("Error initializing regexp " + err.Error())
	}
}

func ValidLen(val string, field string, min int, max int) error {
	if len(val) < min || len(val) > max {
		return errors.New(field + " Must Be Between " + strconv.Itoa(min) + " And " + strconv.Itoa(max) + " Characters")
	}
	return nil
}

/*func ValidName(val *string, field string, name string, errors *binding.Errors) bool {

	return ValidLen(*val, field, name, 2, 50, errors)
}*/

func ValidPhone(val string) (string, error) {

	re := regexp.MustCompile("[^0-9e]")

	val = re.ReplaceAllString(val, "")

	return val, ValidLen(val, "Phone Number", 10, 11)
}

/*func ValidEmail(val *string, field string, name string, errors *binding.Errors) bool {

	*val = strings.ToLower(*val)

	if !ValidLen(*val, field, name, 3, 256, errors) {
		return false
	} else if matched := EmailRegexp.MatchString(*val); matched == false {
		errors.Add([]string{name}, "", "Invalid Email Address")
		return false
	}

	return true
}

func ValidPassword(val *string, field string, name string, errors *binding.Errors) bool {

	return ValidLen(*val, field, name, 8, 256, errors)
}*/
