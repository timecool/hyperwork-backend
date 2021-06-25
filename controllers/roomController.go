package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
	"strconv"
	"time"
	"timecool/hyperwork/database"
	"timecool/hyperwork/handler"
	"timecool/hyperwork/models"
)

var roomCollection *mongo.Collection

func initRoomCollection() {
	database.Ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
	roomCollection = database.DB.Collection("rooms")
}

func CreateRoom(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Create Room")
	w.Header().Add("content-type", "application/json")

	var room models.Room
	err := json.NewDecoder(r.Body).Decode(&room)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	initRoomCollection()
	room.UUID = uuid.New().String()

	//Save User in Collection User
	result, _ := roomCollection.InsertOne(database.Ctx, room)
	json.NewEncoder(w).Encode(result)
}

func DeleteRoom(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Delete Room")
	w.Header().Add("content-type", "application/json")

	params := mux.Vars(r)
	roomId, _ := params["uuid"]

	initRoomCollection()
	initReservationCollection()
	result, err := roomCollection.DeleteOne(database.Ctx, bson.M{"_id": roomId})
	reservationCollection.DeleteMany(database.Ctx, bson.M{"room_uuid": roomId})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(result)
}

func GetRoom(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get Room")
	w.Header().Add("content-type", "application/json")
	params := mux.Vars(r)
	roomId, _ := params["uuid"]

	result, isFind, err := GetRoomById(roomId)
	if err != nil {
		handler.HttpErrorResponse(w, http.StatusBadGateway, err.Error())
		return
	}
	if !isFind {
		handler.HttpErrorResponse(w, http.StatusNotFound, "Room Not Found")
		return
	}
	json.NewEncoder(w).Encode(result)
}

func GetRooms(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get Rooms")
	w.Header().Add("content-type", "application/json")

	var size int64
	var page int64
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
		filter = bson.M{"name": regex}
	}

	var result *mongo.Cursor
	var err error
	skip := page * (size - 1)

	initRoomCollection()
	if page != -1 && size != -1 {
		result, err = roomCollection.Find(
			database.Ctx, filter,
			options.Find().SetProjection(bson.M{"workspaces": 0, "specification": 0}).SetSort(bson.M{"name": 1}).SetLimit(size).SetSkip(skip))
	} else {
		result, err = roomCollection.Find(
			database.Ctx, filter,
			options.Find().SetProjection(bson.M{"workspaces": 0, "specification": 0}).SetSort(bson.M{"name": 1}))
	}
	// find room on UUID
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var rooms []models.Room
	collectionLength, _ := roomCollection.CountDocuments(database.Ctx, filter)
	if err := result.All(database.Ctx, &rooms); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var paging models.PagingRoom
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
	paging.Rooms = rooms

	json.NewEncoder(w).Encode(paging)

}

func UpdateRoom(w http.ResponseWriter, r *http.Request) {
	fmt.Println("UpdateWorkspaceMap")
	w.Header().Add("content-type", "application/json")

	params := mux.Vars(r)
	roomUUID, _ := params["uuid"]
	room, isRoomIdValid, _ := GetRoomById(roomUUID)
	if !isRoomIdValid {
		handler.HttpErrorResponse(w, http.StatusNotFound, "Room was not found")
		return
	}
	var newRoom models.Room

	err := json.NewDecoder(r.Body).Decode(&newRoom)
	if err != nil {
		handler.HttpErrorResponse(w, http.StatusBadGateway, err.Error())
		return
	}

	for i, element := range newRoom.Workspaces {
		if element.UUID == "" {
			newRoom.Workspaces[i].UUID = uuid.New().String()
		}
	}
	initRoomCollection()

	update := bson.D{{"$set", bson.M{
		"name":          newRoom.Name,
		"description":   newRoom.Description,
		"workspaces":    newRoom.Workspaces,
		"specification": newRoom.Specification,
	}}}

	//Update Room
	_, err = roomCollection.UpdateByID(database.Ctx, room.UUID, update)
	if err != nil {
		handler.HttpErrorResponse(w, http.StatusBadGateway, err.Error())
		return
	}
	json.NewEncoder(w).Encode(newRoom)

}

func GetRoomById(id string) (models.Room, bool, error) {
	initRoomCollection()
	var room models.Room
	// find room on UUID
	err := roomCollection.FindOne(database.Ctx, bson.M{"_id": id}).Decode(&room)
	return room, len(room.UUID) != 0, err
}

func getRoomAndWorkspaceName(roomUUID string, workspaceUUID string) (string, string) {
	var room models.Room
	initRoomCollection()
	// find room on UUID
	err := roomCollection.FindOne(database.Ctx, bson.M{"_id": roomUUID, "workspaces._id": workspaceUUID},
		options.FindOne().SetProjection(bson.M{"workspaces.name": 1, "name": 1})).Decode(&room)
	if err != nil {
		return "", ""
	}
	if room.Workspaces == nil {
		return "", ""
	}
	return room.Name, room.Workspaces[0].Name
}
