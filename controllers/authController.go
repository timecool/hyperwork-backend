package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
	"timecool/hyperwork/handler"
	"timecool/hyperwork/models"
	"timecool/hyperwork/util"
)

func Login(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Login")
	w.Header().Add("content-type", "application/json")

	var login models.User
	// decode body to login
	err := json.NewDecoder(r.Body).Decode(&login)
	if err != nil {
		handler.HttpErrorResponse(w, http.StatusUnauthorized, err.Error())
		return
	}

	userInDatabase, isEmailSet, err := GetUserByEmail(login.Email)

	if !isEmailSet {
		// email is wrong
		// in order not to give information what is wrong a general statement is returned
		handler.HttpErrorResponse(w, http.StatusBadRequest, "User or Password wrong")
		return
	}
	if err != nil {
		handler.HttpErrorResponse(w, http.StatusBadGateway, err.Error())
		return
	}
	// check password
	if err := bcrypt.CompareHashAndPassword([]byte(userInDatabase.Password), []byte(login.Password)); err != nil {
		// password is wrong
		// in order not to give information what is wrong a general statement is returned
		handler.HttpErrorResponse(w, http.StatusBadRequest, "User or Password wrong")
		return
	}
	if userInDatabase.UserRole == "none" {
		// user is not activated
		handler.HttpErrorResponse(w, http.StatusUnauthorized, "User not yet activated")
		return
	}

	// create a jwt
	token, err := CreateToken(userInDatabase)
	if err != nil {
		handler.HttpErrorResponse(w, http.StatusBadRequest, err.Error())
	}

	// create cookie with jwt for 2 hours
	cookie := http.Cookie{
		Name:     "jwt",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(time.Hour * 2),
		HttpOnly: true,
	}

	// set cookie in client
	http.SetCookie(w, &cookie)

	// send user without passwort to client
	userInDatabase.Password = ""
	json.NewEncoder(w).Encode(userInDatabase)

}

// parm : user
// return jwt with user datas in claims
func CreateToken(user models.User) (string, error) {
	var err error
	// set new claims
	atClaims := jwt.MapClaims{}
	atClaims["role"] = user.UserRole
	atClaims["uuid"] = user.UUID
	atClaims["name"] = user.Name
	atClaims["email"] = user.Email
	atClaims["exp"] = time.Now().Add(time.Hour * 2)

	// creating access token with claims
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, err := at.SignedString([]byte(util.GetEnvVariable("SECRETKEY")))
	if err != nil {
		return "", err
	}

	return token, nil
}

// delete cookie with jwt
func Logout(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Logout")
	// create jwt cookie which expires in the past
	cookie := http.Cookie{
		Name:     "jwt",
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-time.Hour * 2),
		HttpOnly: true,
	}

	// set cookie in client
	http.SetCookie(w, &cookie)
}
