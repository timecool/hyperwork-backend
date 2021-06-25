package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strconv"
	"time"
	"timecool/hyperwork/database"
	"timecool/hyperwork/handler"
	"timecool/hyperwork/models"
	"timecool/hyperwork/util"
)

var usersCollection *mongo.Collection

func initUserCollection() {
	database.Ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
	usersCollection = database.DB.Collection("users")
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Create User")
	w.Header().Add("content-type", "application/json")
	var user models.User
	//Decode body to user
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	initUserCollection()
	_, isEmailSet, _ := GetUserByEmail(user.Email)
	//if user exists with email then UUID length != 0 and return handler
	if isEmailSet {
		http.Error(w, "Email already exists", http.StatusBadRequest)
		return
	}

	//Hash and Salt the password
	hashPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), 14)
	user.Password = string(hashPassword)
	user.UserRole = models.RoleNone
	//Create uuid UserID
	user.UUID = uuid.New().String()

	//Save User in Collection User
	result, _ := usersCollection.InsertOne(database.Ctx, user)
	json.NewEncoder(w).Encode(result)

}
func GetCurrentUser(r *http.Request) (models.User, error) {
	cookie, err := r.Cookie("jwt")
	var user models.User
	if err != nil {
		return user, err
	}
	token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
		//Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(util.GetEnvVariable("SECRETKEY")), nil

	})
	if err != nil {
		return user, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if ok && token.Valid {
		user = models.User{
			UUID:     claims["uuid"].(string),
			Name:     claims["name"].(string),
			UserRole: models.Role(claims["role"].(string)),
			Email:    claims["email"].(string),
		}

		return user, nil
	}
	return user, nil
}

func GetUserByToken(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get User from Token")
	w.Header().Add("content-type", "application/json")
	user, err := GetCurrentUser(r)
	//Decode body to user
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func GetUsers(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get Users")
	w.Header().Add("content-type", "application/json")
	var size int64
	var page int64
	active := r.URL.Query().Get("active")
	pageString := r.URL.Query().Get("page")
	sizeString := r.URL.Query().Get("size")
	search := r.URL.Query().Get("search")

	if pageString == "" && sizeString == "" {
		// size and page not set
		// set to -1 for error handling
		size = -1
		page = -1
	} else {
		tmpSize, _ := strconv.Atoi(sizeString)
		tmpPage, _ := strconv.Atoi(pageString)
		page = int64(tmpPage)
		size = int64(tmpSize)
	}
	var filter bson.M
	if search != "" {
		// if search => search in the whole collection
		regex := bson.M{"$regex": primitive.Regex{Pattern: search, Options: "i"}}
		filter = bson.M{"$or": []interface{}{
			bson.M{"name": regex},
			bson.M{"email": regex},
		}}
	} else if active == "true" {
		//if ture show only active users
		filter = bson.M{
			"$or": []interface{}{
				bson.M{"role": models.RoleMember},
				bson.M{"role": models.RoleAdmin},
			}}
	} else {
		filter = bson.M{"role": models.RoleNone}
	}
	var result *mongo.Cursor
	var err error
	skip := page * (size - 1)
	initUserCollection()
	if page != -1 && size != -1 {
		result, err = usersCollection.Find(database.Ctx, filter,
			options.Find().SetProjection(bson.M{"password": 0}).SetLimit(size).SetSkip(skip))
	} else {
		result, err = usersCollection.Find(database.Ctx, filter,
			options.Find().SetProjection(bson.M{"password": 0}))
	}
	// get Users form Collection
	if err != nil {
		handler.HttpErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	var users []models.User
	collectionLength, _ := usersCollection.CountDocuments(database.Ctx, filter)
	if err := result.All(database.Ctx, &users); err != nil {
		handler.HttpErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	var paging models.PagingUser
	if page != -1 && size != -1 {
		//Calculations paging
		stepsFloat := float64(collectionLength) / float64(size)
		stepsInt := collectionLength / size
		var steps int64
		if stepsFloat > float64(stepsInt) {
			steps = stepsInt + 1
		} else {
			steps = stepsInt
		}
		paging.PagingInfo.PageNumber = page
		paging.PagingInfo.Steps = steps
	}
	paging.PagingInfo.TotalItems = collectionLength
	paging.Users = users
	json.NewEncoder(w).Encode(paging)

}

func SetRole(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Update Role")
	w.Header().Add("content-type", "application/json")

	params := mux.Vars(r)
	userUUID, _ := params["uuid"]
	var user struct {
		UserRole models.Role `json:"role" bson:"role"`
	}

	//Decode body to user
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		handler.HttpErrorResponse(w, http.StatusBadGateway, err.Error())
		return
	}
	initUserCollection()
	set := bson.M{"$set": bson.M{"role": user.UserRole}}
	result, _ := usersCollection.UpdateByID(database.Ctx, userUUID, set)
	json.NewEncoder(w).Encode(result)
}
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Delete User")
	w.Header().Add("content-type", "application/json")
	params := mux.Vars(r)
	userUUID, err := params["uuid"]

	if !err {
		handler.HttpErrorResponse(w, http.StatusBadGateway, "uuid not set")
		return
	}

	initUserCollection()

	result, err2 := usersCollection.DeleteOne(database.Ctx, bson.M{"_id": userUUID})
	if err2 != nil {
		handler.HttpErrorResponse(w, http.StatusBadGateway, err2.Error())
		return
	}
	json.NewEncoder(w).Encode(result)
}

func GetUserByEmail(email string) (models.User, bool, error) {
	initUserCollection()
	var user models.User
	// find user with email
	err := usersCollection.FindOne(database.Ctx, bson.M{"email": email}).Decode(&user)

	return user, len(user.UUID) != 0, err
}
