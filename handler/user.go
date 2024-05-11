// Package handler of the example application
package handler

import (
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pilinux/argon2"
	"github.com/pilinux/crypt"
	log "github.com/sirupsen/logrus"

	"github.com/tinkerbaj/gintemp/config"
	"github.com/tinkerbaj/gintemp/database"
	"github.com/tinkerbaj/gintemp/lib"
	"github.com/tinkerbaj/gintemp/lib/middleware"
	"github.com/tinkerbaj/gintemp/service"

	// model "github.com/tinkerbaj/gintemp/database/model"

	"github.com/tinkerbaj/gintemp/database/model"
)

// GetUsers handles jobs for controller.GetUsers
func GetUsers() (httpResponse model.HTTPResponse, httpStatusCode int) {
	db := database.GetDB()
	users := []model.User{}
	posts := []model.Post{}
	hobbies := []model.Hobby{}

	if err := db.Find(&users).Error; err != nil {
		log.WithError(err).Error("error code: 1101")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	if len(users) == 0 {
		httpResponse.Message = "no user found"
		httpStatusCode = http.StatusNotFound
		return
	}

	for j, user := range users {
		db.Model(&posts).Where("id_user = ?", user.ID).Find(&posts)
		users[j].Posts = posts

		db.Model(&hobbies).Joins("JOIN user_hobbies ON user_hobbies.hobby_hobby_id=hobbies.hobby_id").
			Joins("JOIN users ON users.user_id=user_hobbies.user_user_id").
			Where("users.user_id = ?", user.ID).
			Find(&hobbies)
		users[j].Hobbies = hobbies
	}

	httpResponse.Message = users
	httpStatusCode = http.StatusOK
	return
}

// GetUser handles jobs for controller.GetUser
func GetUser(id string) (httpResponse model.HTTPResponse, httpStatusCode int) {
	db := database.GetDB()
	user := model.User{}
	posts := []model.Post{}
	hobbies := []model.Hobby{}

	if err := db.Where("user_id = ?", id).First(&user).Error; err != nil {
		httpResponse.Message = "user not found"
		httpStatusCode = http.StatusNotFound
		return
	}

	db.Model(&posts).Where("id_user = ?", user.ID).Find(&posts)
	user.Posts = posts

	db.Model(&hobbies).Joins("JOIN user_hobbies ON user_hobbies.hobby_hobby_id=hobbies.hobby_id").
		Joins("JOIN users ON users.user_id=user_hobbies.user_user_id").
		Where("users.user_id = ?", user.ID).
		Find(&hobbies)
	user.Hobbies = hobbies

	httpResponse.Message = user
	httpStatusCode = http.StatusOK
	return
}

// CreateUser handles jobs for controller.CreateUser
func CreateUser(user model.User) (httpResponse model.HTTPResponse, httpStatusCode int) {
	db := database.GetDB()
	userFinal := model.User{}

	// user must not be able to manipulate all fields
	// authFinal := new(model.User)
	// userFinal.Email = user.Email
	// userFinal.Password = user.Password

	// email validation
	if !lib.ValidateEmail(user.Email) {
		httpResponse.Message = "wrong email address"
		httpStatusCode = http.StatusBadRequest
		return
	}

	// for backward compatibility
	// email must be unique
	err := db.Where("email = ?", user.Email).First(&userFinal).Error
	if err != nil {
		if err.Error() != database.RecordNotFound {
			// db read error
			log.WithError(err).Error("error code: 1002.1")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}
	}
	if err == nil {
		httpResponse.Message = "email already registered"
		httpStatusCode = http.StatusBadRequest
		return
	}

	// downgrade must be avoided to prevent creating duplicate accounts
	// valid: non-encryption mode -> upgrade to encryption mode
	// invalid: encryption mode -> downgrade to non-encryption mode
	if !config.IsCipher() {
		err := db.Where("email_hash IS NOT NULL AND email_hash != ?", "").First(&user).Error
		if err != nil {
			if err.Error() != database.RecordNotFound {
				// db read error
				log.WithError(err).Error("error code: 1002.2")
				httpResponse.Message = "internal server error"
				httpStatusCode = http.StatusInternalServerError
				return
			}
		}
		if err == nil {
			e := errors.New("check env: ACTIVATE_CIPHER")
			log.WithError(e).Error("error code: 1002.3")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}
	}

	// generate a fixed-sized BLAKE2b-256 hash of the email, used for auth purpose
	// when encryption at rest is used
	if config.IsCipher() {
		var err error

		// hash of the email in hexadecimal string format
		emailHash, err := service.CalcHash(
			[]byte(user.Email),
			config.GetConfig().Security.Blake2bSec,
		)
		if err != nil {
			log.WithError(err).Error("error code: 1001.1")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}
		userFinal.EmailHash = hex.EncodeToString(emailHash)

		// email must be unique
		err = db.Where("email_hash = ?", userFinal.EmailHash).First(&user).Error
		if err != nil {
			if err.Error() != database.RecordNotFound {
				// db read error
				log.WithError(err).Error("error code: 1002.4")
				httpResponse.Message = "internal server error"
				httpStatusCode = http.StatusInternalServerError
				return
			}
		}
		if err == nil {
			httpResponse.Message = "email already registered"
			httpStatusCode = http.StatusBadRequest
			return
		}
	}

	// send a verification email if required by the application
	emailDelivered, err := service.SendEmail(userFinal.Email, model.EmailTypeVerifyEmailNewAcc)
	if err != nil {
		log.WithError(err).Error("error code: 1002.5")
		httpResponse.Message = "email delivery service failed"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	if emailDelivered {
		userFinal.VerifyEmail = model.EmailNotVerified
	}

	// encryption at rest for user email, mainly needed by system in future
	// to send verification or password recovery emails
	if config.IsCipher() {
		// encrypt the email
		cipherEmail, nonce, err := crypt.EncryptChacha20poly1305(
			config.GetConfig().Security.CipherKey,
			user.Email,
		)
		if err != nil {
			log.WithError(err).Error("error code: 1001.2")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}

		// save email only in ciphertext
		userFinal.Email = ""
		userFinal.EmailCipher = hex.EncodeToString(cipherEmail)
		userFinal.EmailNonce = hex.EncodeToString(nonce)
	}

	configSecurity := config.GetConfig().Security

	// check password length
	// if more checks are required i.e. password pattern,
	// add all conditions here
	if len(user.Password) < configSecurity.UserPassMinLength {
		log.WithError(err).Error("error code: 1008.2")
		passlen := strconv.Itoa(configSecurity.UserPassMinLength)

		httpResponse.Message = "Password must be at least " + passlen + " long"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	user.Email = strings.TrimSpace(user.Email)

	config := lib.HashPassConfig{
		Memory:      configSecurity.HashPass.Memory,
		Iterations:  configSecurity.HashPass.Iterations,
		Parallelism: configSecurity.HashPass.Parallelism,
		SaltLength:  configSecurity.HashPass.SaltLength,
		KeyLength:   configSecurity.HashPass.KeyLength,
	}
	pass, err := lib.HashPass(config, user.Password, configSecurity.HashSec)
	if err != nil {
		log.WithError(err).Error("error code: 1013.2")
		httpStatusCode = http.StatusInternalServerError
		return
	}
	user.Password = pass

	// one unique email for each account
	tx := db.Begin()
	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("error code: 1001.3")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	tx.Commit()

	httpResponse.Message = user
	httpStatusCode = http.StatusCreated
	return

	// user must not be able to manipulate all fields

	// tx := db.Begin()
	// if err := tx.Create(&userFinal).Error; err != nil {
	// 	tx.Rollback()
	// 	log.WithError(err).Error("error code: 1111")
	// 	httpResponse.Message = "internal server error"
	// 	httpStatusCode = http.StatusInternalServerError
	// 	return
	// }
	// tx.Commit()

	// httpResponse.Message = userFinal
	// httpStatusCode = http.StatusCreated
	// return
}

// UpdateUser handles jobs for controller.UpdateUser
func UpdateUser(userIDAuth uint64, user model.User) (httpResponse model.HTTPResponse, httpStatusCode int) {
	db := database.GetDB()
	userFinal := model.User{}

	// does the user have an existing profile
	if err := db.Where("ID = ?", userIDAuth).First(&userFinal).Error; err != nil {
		httpResponse.Message = "no user profile found"
		httpStatusCode = http.StatusNotFound
		return
	}

	// user must not be able to manipulate all fields
	userFinal.UpdatedAt = time.Now()
	userFinal.FirstName = user.FirstName
	userFinal.LastName = user.LastName

	tx := db.Begin()
	if err := tx.Save(&userFinal).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("error code: 1121")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	tx.Commit()

	httpResponse.Message = userFinal
	httpStatusCode = http.StatusOK
	return
}

// AddHobby handles jobs for controller.AddHobby
func AddHobby(userIDAuth uint64, hobby model.Hobby) (httpResponse model.HTTPResponse, httpStatusCode int) {
	db := database.GetDB()
	user := model.User{}
	hobbyNew := model.Hobby{}
	hobbyFound := 0 // default (do not create new hobby) = 0, create new hobby = 1

	// does the user have an existing profile
	if err := db.Where("ID = ?", userIDAuth).First(&user).Error; err != nil {
		httpResponse.Message = "no user profile found"
		httpStatusCode = http.StatusForbidden
		return
	}

	if err := db.Where("hobby = ?", hobby.Hobby).First(&hobbyNew).Error; err != nil {
		hobbyFound = 1 // create new hobby
	}

	if hobbyFound == 1 {
		hobbyNew.Hobby = hobby.Hobby
		tx := db.Begin()
		if err := tx.Create(&hobbyNew).Error; err != nil {
			tx.Rollback()
			log.WithError(err).Error("error code: 1131")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}
		tx.Commit()
		hobbyFound = 0
	}

	if hobbyFound == 0 {
		user.Hobbies = append(user.Hobbies, hobbyNew)
		tx := db.Begin()
		if err := tx.Save(&user).Error; err != nil {
			tx.Rollback()
			log.WithError(err).Error("error code: 1132")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}
		tx.Commit()
	}

	httpResponse.Message = user
	httpStatusCode = http.StatusOK
	return
}

// -+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-
// CreateUserAuth receives tasks from controller.CreateUserAuth.
// After email validation, it creates a new user account. It
// supports both the legacy way of saving user email in plaintext
// and the recommended way of applying encryption at rest.
// func CreateUserAuth(auth model.User) (httpResponse model.HTTPResponse, httpStatusCode int) {

// }

// UpdateEmail receives tasks from controller.UpdateEmail.
//
// step 1: validate email format
//
// step 2: verify that this email is not registered to anyone
//
// step 3: load user credentials
//
// step 4: verify user password
//
// step 5: calculate hash of the new email
//
// step 6: read 'temp_emails' table
//
// step 7: verify that this is not a repeated request for the same email
//
// step 8: populate model with data to be processed in database
//
// step 9: send a verification email if required by the app
func UpdateEmail(claims middleware.MyCustomClaims, req model.TempEmail) (httpResponse model.HTTPResponse, httpStatusCode int) {
	// check auth validity
	// ok := service.ValidateUserID(claims.UserID)
	// if !ok {
	// 	httpResponse.Message = "access denied"
	// 	httpStatusCode = http.StatusUnauthorized
	// 	return
	// }

	// step 1: validate email format
	req.Email = strings.TrimSpace(req.Email)
	if !lib.ValidateEmail(req.Email) {
		httpResponse.Message = "wrong email address"
		httpStatusCode = http.StatusBadRequest
		return
	}

	// step 2: verify that this email is not registered to anyone
	_, err := service.GetUserByEmail(req.Email, false)
	if err != nil {
		if err.Error() != database.RecordNotFound {
			// db read error
			log.WithError(err).Error("error code: 1003.21")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}
	}
	if err == nil {
		httpResponse.Message = "email already registered"
		httpStatusCode = http.StatusBadRequest
		return
	}
	// ok: email is not registered yet, continue...

	// db connection
	db := database.GetDB()

	// step 3: load user credentials
	auth := model.User{}
	err = db.Where("ID = ?", claims.UserID).First(&auth).Error
	if err != nil {
		// most likely db read error
		log.WithError(err).Error("error code: 1003.31")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	// app settings
	configSecurity := config.GetConfig().Security

	// step 4: verify user password
	verifyPass, err := argon2.ComparePasswordAndHash(req.Password, configSecurity.HashSec, auth.Password)
	if err != nil {
		log.WithError(err).Error("error code: 1003.41")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	if !verifyPass {
		httpResponse.Message = "wrong password"
		httpStatusCode = http.StatusBadRequest
		return
	}

	// step 5: calculate hash of the new email
	emailHash, err := service.CalcHash(
		[]byte(req.Email),
		config.GetConfig().Security.Blake2bSec,
	)
	if err != nil {
		log.WithError(err).Error("error code: 1003.51")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	// step 6: read 'temp_emails' table
	tEmailDB := model.TempEmail{}
	err = db.Where("ID = ?", claims.UserID).First(&tEmailDB).Error
	if err != nil {
		if err.Error() != database.RecordNotFound {
			// db read error
			log.WithError(err).Error("error code: 1003.61")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}
		// this user has no previous pending request to update email
	}

	// this user has a pending request to update the email
	if err == nil {
		// step 7: verify that this is not a repeated request for the same email

		// plaintext
		if tEmailDB.Email != "" {
			if tEmailDB.Email == req.Email {
				httpResponse.Message = "please verify the new email"
				httpStatusCode = http.StatusBadRequest
				return
			}
		}

		// encryption at rest
		if tEmailDB.Email == "" {
			if tEmailDB.EmailHash == hex.EncodeToString(emailHash) {
				httpResponse.Message = "please verify the new email"
				httpStatusCode = http.StatusBadRequest
				return
			}
		}
	}

	// step 8: populate model with data to be processed in database
	timeNow := time.Now()

	// create new data
	if tEmailDB.ID == 0 {
		tEmailDB.CreatedAt = timeNow
		tEmailDB.IDAuth = claims.UserID
	}

	tEmailDB.UpdatedAt = timeNow

	// plaintext
	if !config.IsCipher() {
		tEmailDB.Email = req.Email
		tEmailDB.EmailCipher = ""
		tEmailDB.EmailNonce = ""
		tEmailDB.EmailHash = ""
	}

	// encryption at rest
	if config.IsCipher() {
		tEmailDB.Email = ""

		// encrypt the email
		cipherEmail, nonce, err := crypt.EncryptChacha20poly1305(
			config.GetConfig().Security.CipherKey,
			req.Email,
		)
		if err != nil {
			log.WithError(err).Error("error code: 1003.81")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}

		// save email only in ciphertext
		tEmailDB.EmailCipher = hex.EncodeToString(cipherEmail)
		tEmailDB.EmailNonce = hex.EncodeToString(nonce)
		tEmailDB.EmailHash = hex.EncodeToString(emailHash)
	}

	// step 9: send a verification email if required by the application
	emailDelivered, err := service.SendEmail(req.Email, model.EmailTypeVerifyUpdatedEmail)
	if err != nil {
		log.WithError(err).Error("error code: 1003.91")
		httpResponse.Message = "email delivery service failed"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	// verification code sent
	if emailDelivered {
		tx := db.Begin()
		if err := tx.Save(&tEmailDB).Error; err != nil {
			tx.Rollback()
			log.WithError(err).Error("error code: 1003.92")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}
		tx.Commit()

		httpResponse.Message = "verification email delivered"
		httpStatusCode = http.StatusOK
	}

	// verification code not required, update email immediately
	if !emailDelivered {
		auth.UpdatedAt = timeNow
		auth.Email = tEmailDB.Email
		auth.EmailCipher = tEmailDB.EmailCipher
		auth.EmailNonce = tEmailDB.EmailNonce
		auth.EmailHash = tEmailDB.EmailHash

		tx := db.Begin()
		if err := tx.Save(&auth).Error; err != nil {
			tx.Rollback()
			log.WithError(err).Error("error code: 1003.93")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}
		tx.Commit()

		httpResponse.Message = "email updated"
		httpStatusCode = http.StatusOK
	}

	return
}
