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
	// connection to users collection
	// save in global package variable
	usersCollection = database.DB.Collection("users")
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Create User")
	w.Header().Add("content-type", "application/json")
	var user models.User

	// decode body to user
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		handler.HttpErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	initUserCollection()

	_, isEmailSet, _ := GetUserByEmail(user.Email)
	//if user exists with email then UUID length != 0 and return handler
	if isEmailSet {
		handler.HttpErrorResponse(w, http.StatusBadRequest, "Email already exists")
		return
	}

	// hash and salt the password
	hashPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), 14)
	user.Password = string(hashPassword)
	user.UserRole = models.RoleNone

	// create uuid
	user.UUID = uuid.New().String()

	// save user in collection
	result, _ := usersCollection.InsertOne(database.Ctx, user)

	// send datas to client
	json.NewEncoder(w).Encode(result)

}
func GetCurrentUser(r *http.Request) (models.User, error) {
	cookie, err := r.Cookie("jwt")
	var user models.User
	if err != nil {
		return user, err
	}
	token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
		// make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(util.GetEnvVariable("SECRETKEY")), nil

	})
	if err != nil {
		return user, err
	}

	// get token Claims
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

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Update User")
	w.Header().Add("content-type", "application/json")

	// get uuid from url
	params := mux.Vars(r)
	userUUID, _ := params["uuid"]

	// get User by UUID
	oldUser, isFind, _ := GetUserById(userUUID)

	if !isFind {
		handler.HttpErrorResponse(w, http.StatusNotFound, "User not found")
		return
	}
	var user models.User

	// decode body to user
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		handler.HttpErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	if oldUser.Email != user.Email {
		// check is email in collection
		_, emailFound, _ := GetUserByEmail(user.Email)
		if emailFound {
			// email is in collection
			handler.HttpErrorResponse(w, http.StatusBadRequest, "Email is already in use")
			return
		}
	}
	// update user bson
	update := bson.D{{"$set", bson.M{"name": user.Name, "email": user.Email, "role": user.UserRole}}}

	initUserCollection()

	if user.Password != "" {
		// if the password is set it must be hashed first
		hashPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), 14)
		user.Password = string(hashPassword)
		update = bson.D{{"$set", bson.M{"name": user.Name, "password": user.Password, "email": user.Email, "role": user.UserRole}}}
	}

	// check User role, is it right
	if user.UserRole == models.RoleNone || user.UserRole == models.RoleMember || user.UserRole == models.RoleAdmin {
		// user Update by ID
		_, err := usersCollection.UpdateByID(database.Ctx, userUUID, update)
		if err != nil {
			handler.HttpErrorResponse(w, http.StatusBadRequest, err.Error())
		}
		// send new user to client without password
		user.Password = ""
		json.NewEncoder(w).Encode(user)
		return
	} else {
		handler.HttpErrorResponse(w, http.StatusBadRequest, "Wrong Role")
		return
	}
}
func GetUserByToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json")

	user, err := GetCurrentUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// send datas to client
	json.NewEncoder(w).Encode(user)
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get User")
	w.Header().Add("content-type", "application/json")

	// get uuid from url
	params := mux.Vars(r)
	userId, _ := params["uuid"]

	// get user
	result, isFind, err := GetUserById(userId)
	if err != nil {
		handler.HttpErrorResponse(w, http.StatusBadGateway, err.Error())
		return
	}
	if !isFind {
		handler.HttpErrorResponse(w, http.StatusNotFound, "User Not Found")
		return
	}

	// send user to client without password
	result.Password = ""
	json.NewEncoder(w).Encode(result)
}
func GetUsers(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get Users")
	w.Header().Add("content-type", "application/json")
	var size int64
	var page int64

	// fetches the parameters from the url
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
	// calc skip value
	skip := page * (size - 1)

	initUserCollection()

	if page != -1 && size != -1 {
		// page and size is set => find in collection user
		result, err = usersCollection.Find(database.Ctx, filter,
			options.Find().SetProjection(bson.M{"password": 0}).SetLimit(size).SetSkip(skip))
	} else {
		// get users without page and size
		result, err = usersCollection.Find(database.Ctx, filter,
			options.Find().SetProjection(bson.M{"password": 0}))
	}
	if err != nil {
		handler.HttpErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	var users []models.User

	// count collection length for total items
	collectionLength, _ := usersCollection.CountDocuments(database.Ctx, filter)
	if err := result.All(database.Ctx, &users); err != nil {
		handler.HttpErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// create paging information
	var paging models.PagingUser
	if page != -1 && size != -1 {
		// calculations paging
		stepsFloat := float64(collectionLength) / float64(size)
		stepsInt := collectionLength / size
		var steps int64
		// round value high if stepfloat it is a comma number
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

	// send datas to client
	json.NewEncoder(w).Encode(paging)

}

func SetRole(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Update Role")
	w.Header().Add("content-type", "application/json")

	// get uuid from url
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

	// check role
	if user.UserRole == models.RoleNone || user.UserRole == models.RoleMember || user.UserRole == models.RoleAdmin {
		set := bson.M{"$set": bson.M{"role": user.UserRole}}
		result, _ := usersCollection.UpdateByID(database.Ctx, userUUID, set)

		// send datas to client
		json.NewEncoder(w).Encode(result)
	} else {
		handler.HttpErrorResponse(w, http.StatusBadGateway, "Wrong Role")
		return
	}

}
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Delete User")
	w.Header().Add("content-type", "application/json")

	// get uuid from url
	params := mux.Vars(r)
	userUUID, err := params["uuid"]

	if !err {
		handler.HttpErrorResponse(w, http.StatusBadGateway, "uuid not set")
		return
	}

	initUserCollection()

	// delete user by uuid
	_, err2 := usersCollection.DeleteOne(database.Ctx, bson.M{"_id": userUUID})
	if err2 != nil {
		handler.HttpErrorResponse(w, http.StatusBadGateway, err2.Error())
		return
	}

	// delete all reservation with user_uuid and time in future
	unix := time.Now().Unix()
	filter := bson.M{"$and": []interface{}{
		bson.M{"start_date": bson.M{"$gte": unix}},
		bson.M{"user_uuid": userUUID}}}
	_, err2 = reservationCollection.DeleteMany(database.Ctx, filter)
	if err2 != nil {
		handler.HttpErrorResponse(w, http.StatusBadGateway, err2.Error())
		return
	}
}

func GetUserByEmail(email string) (models.User, bool, error) {
	initUserCollection()
	var user models.User

	// find user with email
	err := usersCollection.FindOne(database.Ctx, bson.M{"email": email}).Decode(&user)

	return user, len(user.UUID) != 0, err
}

// parms: uuid as string
// return user, bool => is user find, error
func GetUserById(id string) (models.User, bool, error) {
	initUserCollection()
	var user models.User
	// find user on UUID
	err := usersCollection.FindOne(database.Ctx, bson.M{"_id": id}).Decode(&user)
	return user, len(user.UUID) != 0, err
}
