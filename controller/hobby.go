package controller

import (
	"github.com/gin-gonic/gin"

	"github.com/tinkerbaj/gintemp/lib/renderer"

	"github.com/tinkerbaj/gintemp/handler"
)

// GetHobbies - GET /hobbies
func GetHobbies(c *gin.Context) {
	resp, statusCode := handler.GetHobbies()

	renderer.Render(c, resp, statusCode)
}
