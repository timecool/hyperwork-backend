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
	// connection to rooms collection
	// save in global package variable
	roomCollection = database.DB.Collection("rooms")
}

func CreateRoom(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Create Room")
	w.Header().Add("content-type", "application/json")

	var room models.Room
	// decode body to room
	err := json.NewDecoder(r.Body).Decode(&room)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	initRoomCollection()

	// create uuid
	room.UUID = uuid.New().String()
	// save room in collection room
	result, _ := roomCollection.InsertOne(database.Ctx, room)

	// send datas to client
	json.NewEncoder(w).Encode(result)
}

func DeleteRoom(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Delete Room")
	w.Header().Add("content-type", "application/json")

	// get uuid from url
	params := mux.Vars(r)
	roomId, _ := params["uuid"]

	initRoomCollection()
	initReservationCollection()

	// delete room by uuid
	result, err := roomCollection.DeleteOne(database.Ctx, bson.M{"_id": roomId})
	// remove all reservation with roomUUID
	reservationCollection.DeleteMany(database.Ctx, bson.M{"room_uuid": roomId})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// send datas to client
	json.NewEncoder(w).Encode(result)
}

func GetRoom(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get Room")
	w.Header().Add("content-type", "application/json")

	// get uuid from url
	params := mux.Vars(r)
	roomId, _ := params["uuid"]

	// get room by uuid
	result, isFind, err := GetRoomById(roomId)
	if err != nil {
		handler.HttpErrorResponse(w, http.StatusBadGateway, err.Error())
		return
	}
	if !isFind {
		handler.HttpErrorResponse(w, http.StatusNotFound, "Room Not Found")
		return
	}

	// send datas to client
	json.NewEncoder(w).Encode(result)
}

func GetRooms(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get Rooms")
	w.Header().Add("content-type", "application/json")

	var size int64
	var page int64
	// fetches the parameters from the url
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

	// calc skip value
	skip := page * (size - 1)

	initRoomCollection()

	if page != -1 && size != -1 {
		// page and size is set => find in collection rooms
		// Settings: show  the room without workspaces and specification
		// Sort by name
		result, err = roomCollection.Find(
			database.Ctx, filter,
			options.Find().SetProjection(bson.M{"workspaces": 0, "specification": 0}).SetSort(bson.M{"name": 1}).SetLimit(size).SetSkip(skip))
	} else {
		// Settings: show  the room without workspaces and specification
		// Sort by name
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

	// count collection length for total items
	collectionLength, _ := roomCollection.CountDocuments(database.Ctx, filter)

	// decode result to room array
	if err := result.All(database.Ctx, &rooms); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// create paging information
	var paging models.PagingRoom
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
	paging.Rooms = rooms

	// send datas to client
	json.NewEncoder(w).Encode(paging)

}

func UpdateRoom(w http.ResponseWriter, r *http.Request) {
	fmt.Println("UpdateWorkspaceMap")
	w.Header().Add("content-type", "application/json")

	// get uuid from url
	params := mux.Vars(r)
	roomUUID, _ := params["uuid"]

	// get room by uuid
	room, isRoomIdValid, _ := GetRoomById(roomUUID)
	if !isRoomIdValid {
		handler.HttpErrorResponse(w, http.StatusNotFound, "Room was not found")
		return
	}
	var newRoom models.Room

	//Decode body to user
	err := json.NewDecoder(r.Body).Decode(&newRoom)
	if err != nil {
		handler.HttpErrorResponse(w, http.StatusBadGateway, err.Error())
		return
	}
	for i, newElement := range newRoom.Workspaces {
		if newElement.UUID == "" {
			// if uuid emty = new room = new uuid
			newRoom.Workspaces[i].UUID = uuid.New().String()
		}
	}

	var deleteList []string
	// check is workspace delete
	// is workspace delete save in deleteList
	for _, oldElement := range room.Workspaces {
		isNotInNew := true
		for _, newElement := range newRoom.Workspaces {
			if newElement.UUID == oldElement.UUID {
				isNotInNew = false
			}
		}
		if isNotInNew {
			deleteList = append(deleteList, oldElement.UUID)
		}
	}

	if len(deleteList) != 0 {
		initReservationCollection()
		// if workspace delete, delete all reservation with workspace_uuid
		_, err := reservationCollection.DeleteMany(database.Ctx, bson.M{"workspace_uuid": bson.M{"$in": deleteList}})
		if err != nil {
			handler.HttpErrorResponse(w, http.StatusBadGateway, err.Error())
			return
		}
	}

	initRoomCollection()

	// update room bson
	update := bson.D{{"$set", bson.M{
		"description":   newRoom.Description,
		"workspaces":    newRoom.Workspaces,
		"specification": newRoom.Specification,
	}}}

	// update room by UUID
	_, err = roomCollection.UpdateByID(database.Ctx, room.UUID, update)
	if err != nil {
		handler.HttpErrorResponse(w, http.StatusBadGateway, err.Error())
		return
	}

	// send datas to client
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
	// find workspace by roomUUID and workspaceUUID
	// Settings: get only workspace name and room name
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
