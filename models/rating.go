package models

import "github.com/jinzhu/gorm"

type Rating struct {
	gorm.Model
	UserID      uint
	RatedUserID uint
	RideID      uint
	Rating      uint8
}

type userRatingResult struct {
	Count uint
	Avg   float64
}

func GetUserRating(db *gorm.DB, id uint) float64 {
	rating := &userRatingResult{}
	db.Model(&Rating{}).Select("avg(rating) AS avg, count(*) AS count").Where("rated_user_id = ?", id).Scan(rating)
	if rating.Count > 0 {
		return rating.Avg
	} else {
		return 5.0
	}
}
