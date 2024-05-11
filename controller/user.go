// Package controller contains all the controllers
// of the application
package controller

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/tinkerbaj/gintemp/config"
	"github.com/tinkerbaj/gintemp/lib/renderer"
	"github.com/tinkerbaj/gintemp/service"

	"github.com/tinkerbaj/gintemp/database/model"
	"github.com/tinkerbaj/gintemp/handler"
)

// GetUsers - GET /users
func GetUsers(c *gin.Context) {
	resp, statusCode := handler.GetUsers()

	renderer.Render(c, resp, statusCode)
}

// GetUser - GET /users/:id
func GetUser(c *gin.Context) {
	id := strings.TrimSpace(c.Params.ByName("id"))

	resp, statusCode := handler.GetUser(id)

	if reflect.TypeOf(resp.Message).Kind() == reflect.String {
		renderer.Render(c, resp, statusCode)
		return
	}

	renderer.Render(c, resp.Message, statusCode)
}

// CreateUser - POST /users
func CreateUser(c *gin.Context) {
	// userIDAuth := c.GetUint("userID")
	user := model.User{}

	// verify that RDBMS is enabled in .env
	if !config.IsRDBMS() {
		renderer.Render(c, gin.H{"message": "relational database not enabled"}, http.StatusNotImplemented)
		return
	}

	// req, _ := io.ReadAll(c.Request.Body)
	// fmt.Println(string(req))

	// bind JSON
	if err := c.ShouldBindJSON(&user); err != nil {
		fmt.Println(err)
		renderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	// resp, statusCode := handler.CreateUserAuth(user)

	// if reflect.TypeOf(resp.Message).Kind() == reflect.String {
	// 	renderer.Render(c, resp, statusCode)
	// 	return
	// }

	// renderer.Render(c, resp.Message, statusCode)

	// bind JSON
	// if err := c.ShouldBindJSON(&user); err != nil {
	// 	renderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
	// 	return
	// }

	resp, statusCode := handler.CreateUser(user)

	if reflect.TypeOf(resp.Message).Kind() == reflect.String {
		renderer.Render(c, resp, statusCode)
		return
	}

	_, errAccessJWT := c.Cookie("accessJWT")
	_, errRefreshJWT := c.Cookie("refreshJWT")
	if errAccessJWT == nil || errRefreshJWT == nil {
		configSecurity := config.GetConfig().Security
		c.SetCookie(
			"accessJWT",
			"",
			-1,
			configSecurity.AuthCookiePath,
			configSecurity.AuthCookieDomain,
			configSecurity.AuthCookieSecure,
			configSecurity.AuthCookieHTTPOnly,
		)
		c.SetCookie(
			"refreshJWT",
			"",
			-1,
			configSecurity.AuthCookiePath,
			configSecurity.AuthCookieDomain,
			configSecurity.AuthCookieSecure,
			configSecurity.AuthCookieHTTPOnly,
		)
	}
	renderer.Render(c, resp.Message, statusCode)
}

// UpdateUser - PUT /users
func UpdateUser(c *gin.Context) {
	userIDAuth := c.GetUint64("userID")
	user := model.User{}

	// bind JSON
	if err := c.ShouldBindJSON(&user); err != nil {
		renderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.UpdateUser(userIDAuth, user)

	if reflect.TypeOf(resp.Message).Kind() == reflect.String {
		renderer.Render(c, resp, statusCode)
		return
	}

	renderer.Render(c, resp.Message, statusCode)
}

// AddHobby - PUT /users/hobbies
func AddHobby(c *gin.Context) {
	// userIDAuth := c.GetUint64("userID")
	hobby := model.Hobby{}

	// bind JSON
	if err := c.ShouldBindJSON(&hobby); err != nil {
		renderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.AddHobby(17, hobby)

	if reflect.TypeOf(resp.Message).Kind() == reflect.String {
		renderer.Render(c, resp, statusCode)
		return
	}

	renderer.Render(c, resp.Message, statusCode)
}

func CreateUserAuth(c *gin.Context) {
	// delete existing auth cookie if present
	_, errAccessJWT := c.Cookie("accessJWT")
	_, errRefreshJWT := c.Cookie("refreshJWT")
	if errAccessJWT == nil || errRefreshJWT == nil {
		configSecurity := config.GetConfig().Security
		c.SetCookie(
			"accessJWT",
			"",
			-1,
			configSecurity.AuthCookiePath,
			configSecurity.AuthCookieDomain,
			configSecurity.AuthCookieSecure,
			configSecurity.AuthCookieHTTPOnly,
		)
		c.SetCookie(
			"refreshJWT",
			"",
			-1,
			configSecurity.AuthCookiePath,
			configSecurity.AuthCookieDomain,
			configSecurity.AuthCookieSecure,
			configSecurity.AuthCookieHTTPOnly,
		)
	}

	// verify that RDBMS is enabled in .env
	if !config.IsRDBMS() {
		renderer.Render(c, gin.H{"message": "relational database not enabled"}, http.StatusNotImplemented)
		return
	}

	auth := model.User{}

	// bind JSON
	if err := c.ShouldBindJSON(&auth); err != nil {
		renderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.CreateUser(auth)

	if reflect.TypeOf(resp.Message).Kind() == reflect.String {
		renderer.Render(c, resp, statusCode)
		return
	}

	renderer.Render(c, resp.Message, statusCode)
}

// UpdateEmail - update existing user email
//
// dependency: relational database, JWT
//
// Accepted JSON payload:
//
// `{"emailNew":"...", "password":"..."}`
func UpdateEmail(c *gin.Context) {
	// verify that RDBMS is enabled in .env
	if !config.IsRDBMS() {
		renderer.Render(c, gin.H{"message": "relational database not enabled"}, http.StatusNotImplemented)
		return
	}

	// verify that JWT service is enabled in .env
	if !config.IsJWT() {
		renderer.Render(c, gin.H{"message": "JWT service not enabled"}, http.StatusNotImplemented)
		return
	}

	// get claims
	claims := service.GetClaims(c)

	req := model.TempEmail{}

	// bind JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		renderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.UpdateEmail(claims, req)

	renderer.Render(c, resp, statusCode)
}
