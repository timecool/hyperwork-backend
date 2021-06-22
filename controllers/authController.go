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
	err := json.NewDecoder(r.Body).Decode(&login)
	if err != nil {
		handler.HttpErrorResponse(w, http.StatusUnauthorized, err.Error())
		return
	}

	userInDatabase, isEmailSet, err := GetUserByEmail(login.Email)

	if !isEmailSet {
		handler.HttpErrorResponse(w, http.StatusBadRequest, "User or Password wrong")
		return
	}
	if err != nil {
		fmt.Println("authController:27 get User by email handler")
		handler.HttpErrorResponse(w, http.StatusBadGateway, err.Error())
		return
	}
	fmt.Println(userInDatabase)
	//Check Password
	if err := bcrypt.CompareHashAndPassword([]byte(userInDatabase.Password), []byte(login.Password)); err != nil {
		handler.HttpErrorResponse(w, http.StatusBadRequest, "User or Password wrong")
		return
	}
	//Create JWT
	token, err := CreateToken(userInDatabase)

	if err != nil {
		handler.HttpErrorResponse(w, http.StatusBadRequest, err.Error())
	}
	//Create Cookie with jwt Token for 2 Hours
	cookie := http.Cookie{
		Name:     "jwt",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(time.Hour * 2),
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
	userInDatabase.Password = ""
	json.NewEncoder(w).Encode(userInDatabase)

}

func CreateToken(user models.User) (string, error) {
	var err error
	//Creating Access Token
	atClaims := jwt.MapClaims{}
	atClaims["role"] = user.UserRole
	atClaims["uuid"] = user.UUID
	atClaims["name"] = user.Name
	atClaims["email"] = user.Email
	atClaims["exp"] = time.Now().Add(time.Hour * 2)
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, err := at.SignedString([]byte(util.GetEnvVariable("SECRETKEY")))
	if err != nil {
		return "", err
	}
	return token, nil
}

//Delete Cookie with jwt Token
func Logout(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Logout")
	cookie := http.Cookie{
		Name:     "jwt",
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-time.Hour * 2),
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)

}
