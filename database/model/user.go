// Package model contains all the models required
// for a functional database management system
// package model

// import (
// 	"time"

// 	"gorm.io/gorm"
// )

// // User model - `users` table
// type User struct {
// 	UserID    uint64         `gorm:"primaryKey" json:"userID,omitempty"`
// 	CreatedAt time.Time      `json:"createdAt,omitempty"`
// 	UpdatedAt time.Time      `json:"updatedAt,omitempty"`
// 	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
// 	FirstName string         `json:"firstName,omitempty"`
// 	LastName  string         `json:"lastName,omitempty"`
// 	IDAuth    uint64         `json:"-"`
// 	Posts     []Post         `gorm:"foreignkey:IDUser;references:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"posts,omitempty"`
// 	Hobbies   []Hobby        `gorm:"many2many:user_hobbies" json:"hobbies,omitempty"`
// }

package model

import (
	// "github.com/google/uuid"
	// "encoding/json"
	// "errors"
	// "strings"

	// "github.com/tinkerbaj/gintemp/config"
	// "github.com/tinkerbaj/gintemp/lib"
	"gorm.io/gorm"
)

// Email verification statuses
const (
	EmailNotVerified       int8 = -1
	EmailVerifyNotRequired int8 = 0
	EmailVerified          int8 = 1
)

// Email type
const (
	EmailTypeVerifyEmailNewAcc  int = 1 // verify email of newly registered user
	EmailTypePassRecovery       int = 2 // password recovery code
	EmailTypeVerifyUpdatedEmail int = 3 // verify request of updating user email
)

// Redis key prefixes
const (
	EmailVerificationKeyPrefix string = "gintemp-email-verification-"
	EmailUpdateKeyPrefix       string = "gintemp-email-update-"
	PasswordRecoveryKeyPrefix  string = "gintemp-pass-recover-"
)

// GetAddress is the helper struct to get the address fields on one place
type Address struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	City    string `json:"city"`
	State   string `json:"state"`
	Zip     string `json:"zip"`
}

type User struct {
	gorm.Model
	FirstName             string  `json:"firstName"`
	LastName              string  `json:"lastName"`
	Posts                 []Post  `json:"posts,omitempty" gorm:"foreignKey:UserID"`
	Hobbies               []Hobby `json:"hobbies,omitempty" gorm:"many2many:user_hobbies;"`
	Name                  string  `json:"name" `
	Username              string  `gorm:"unique" json:"username"`
	Email                 string  `gorm:"unique" json:"email"`
	EmailCipher           string  `json:"-"`
	EmailNonce            string  `json:"-"`
	EmailHash             string  `gorm:"index" json:"-"`
	VerifyEmail           int8    `json:"-"`
	Password              string  `json:"password"`
	EmailVerified         bool    `json:"email_verified"`
	Phone                 string  `json:"phone"`
	PhoneVerified         bool    `json:"phone_verified"`
	Address               string  `json:"address"`
	City                  string  `json:"city"`
	State                 string  `json:"state"`
	Zip                   string  `json:"zip"`
	Latitude              float64 `json:"latitude"`
	Longitude             float64 `json:"longitude"`
	IsShop                bool    `json:"is_shop" gorm:"default:false"`
	IsAdmin               bool    `json:"is_admin" gorm:"default:false"`
	IsActive              bool    `json:"is_active" gorm:"default:false"`
	IsDeleted             bool    `json:"is_deleted"`
	Role                  string  `json:"role" gorm:"default:'user'"`
	Status                string  `json:"status"`
	ProfileImage          string  `json:"profile_image"`
	CoverImage            string  `json:"cover_image"`
	AboutMe               string  `json:"about_me"`
	Facebook              string  `json:"facebook"`
	Twitter               string  `json:"twitter"`
	Instagram             string  `json:"instagram"`
	Google                string  `json:"google"`
	Linkedin              string  `json:"linkedin"`
	Youtube               string  `json:"youtube"`
	Website               string  `json:"website"`
	VerificationToken     string  `json:"verification_token"`
	VerificationExpireAt  int64   `json:"verification_expire_at"`
	ResetPasswordToken    string  `json:"reset_password_token"`
	ResetPasswordExpireAt int64   `json:"reset_password_expire_at"`
}

// GetAddress is the helper function to get the address
func (u *User) GetAddress() Address {
	return Address{
		Name:    u.Name,
		Address: u.Address,
		City:    u.City,
		State:   u.State,
		Zip:     u.Zip,
	}

}

// UnmarshalJSON ...
// func (v *User) UnmarshalJSON(b []byte) error {
// 	aux := struct {
// 		Email    string `json:"email"`
// 		Password string `json:"password"`
// 	}{}
// 	if err := json.Unmarshal(b, &aux); err != nil {
// 		return err
// 	}

// 	configSecurity := config.GetConfig().Security

// 	// check password length
// 	// if more checks are required i.e. password pattern,
// 	// add all conditions here
// 	if len(aux.Password) < configSecurity.UserPassMinLength {
// 		return errors.New("short password")
// 	}

// 	v.Email = strings.TrimSpace(aux.Email)

// 	config := lib.HashPassConfig{
// 		Memory:      configSecurity.HashPass.Memory,
// 		Iterations:  configSecurity.HashPass.Iterations,
// 		Parallelism: configSecurity.HashPass.Parallelism,
// 		SaltLength:  configSecurity.HashPass.SaltLength,
// 		KeyLength:   configSecurity.HashPass.KeyLength,
// 	}
// 	pass, err := lib.HashPass(config, aux.Password, configSecurity.HashSec)
// 	if err != nil {
// 		return err
// 	}
// 	v.Password = pass

// 	return nil
// }

// MarshalJSON ...
// func (v User) MarshalJSON() ([]byte, error) {
// 	aux := struct {
// 		Email  string `json:"email"`
// 	}{
// 		Email:  strings.TrimSpace(v.Email),
// 	}

// 	return json.Marshal(aux)
// }

// AuthPayload - struct to handle all auth data
type AuthPayload struct {
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`

	VerificationCode string `json:"verificationCode,omitempty"`

	OTP string `json:"otp,omitempty"`

	SecretCode  string `json:"secretCode,omitempty"`
	RecoveryKey string `json:"recoveryKey,omitempty"`

	PassNew    string `json:"passNew,omitempty"`
	PassRepeat string `json:"passRepeat,omitempty"`
}

// TempEmail - 'temp_emails' table to hold data temporarily
// during the process of replacing a user's email address
// with a new one
type TempEmail struct {
	gorm.Model
	Email       string `gorm:"index" json:"emailNew"`
	Password    string `gorm:"-" json:"password,omitempty"`
	EmailCipher string `json:"-"`
	EmailNonce  string `json:"-"`
	EmailHash   string `gorm:"index" json:"-"`
	IDAuth      uint   `gorm:"index" json:"-"`
}
