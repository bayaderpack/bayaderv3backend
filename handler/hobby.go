package handler

import (
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/tinkerbaj/gintemp/database"

	"github.com/tinkerbaj/gintemp/database/model"
)

// GetHobbies handles jobs for controller.GetHobbies
func GetHobbies() (httpResponse model.HTTPResponse, httpStatusCode int) {
	db := database.GetDB()
	hobbies := []model.Hobby{}

	if err := db.Find(&hobbies).Error; err != nil {
		log.WithError(err).Error("error code: 1251")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	if len(hobbies) == 0 {
		httpResponse.Message = "no hobby found"
		httpStatusCode = http.StatusNotFound
		return
	}

	httpResponse.Message = hobbies
	httpStatusCode = http.StatusOK
	return
}
