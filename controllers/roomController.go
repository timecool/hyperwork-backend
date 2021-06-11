package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
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
	w.Header().Add("content-type", "application")

	var room models.Room
	err := json.NewDecoder(r.Body).Decode(&room)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	initRoomCollection()
	room.UUID = uuid.New().String()
	room.Delete = false

	//Save User in Collection User
	result, _ := roomCollection.InsertOne(database.Ctx, room)
	json.NewEncoder(w).Encode(result)
}

func DeleteRoom(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Create Room")
	w.Header().Add("content-type", "application")

	params := mux.Vars(r)
	roomId, _ := params["uuid"]

	initRoomCollection()
	opts := options.Update().SetUpsert(true)
	filter := bson.D{{"_id", roomId}}
	update := bson.D{{"$set", bson.D{{"delete", true}}}}

	result, err := roomCollection.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(result)
}

func GetRoom(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get Room")
	w.Header().Add("content-type", "application")
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
	w.Header().Add("content-type", "application")

	initRoomCollection()
	// find room on UUID
	result, err := roomCollection.Find(database.Ctx, bson.M{"delete": false}, options.Find().SetProjection(bson.M{"workspaces": 0, "specification": 0}))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var room []models.Room

	if err := result.All(database.Ctx, &room); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(room)

}

func UpdateRoom(w http.ResponseWriter, r *http.Request) {
	fmt.Println("UpdateWorkspaceMap")
	w.Header().Add("content-type", "application")

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

	var result *mongo.UpdateResult
	update := bson.D{{"$set", bson.M{
		"name":          newRoom.Name,
		"description":   newRoom.Description,
		"workspaces":    newRoom.Workspaces,
		"specification": newRoom.Specification,
	}}}

	//Update Room
	result, err = roomCollection.UpdateByID(database.Ctx, room.UUID, update)
	if err != nil {
		handler.HttpErrorResponse(w, http.StatusBadGateway, err.Error())
		return
	}
	json.NewEncoder(w).Encode(result)

}

func GetRoomById(id string) (models.Room, bool, error) {
	initRoomCollection()
	var room models.Room
	// find room on UUID
	err := roomCollection.FindOne(database.Ctx, bson.M{"_id": id}).Decode(&room)
	return room, len(room.UUID) != 0 && !room.Delete, err
}
