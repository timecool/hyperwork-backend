package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
	"timecool/hyperwork/database"
	"timecool/hyperwork/models"
)

var workspaceCollection *mongo.Collection

func DeleteWorkspace(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Delete Workspace")
	w.Header().Add("content-type", "application")

	params := mux.Vars(r)
	roomId, _ := params["uuid"]

	opts := options.Update().SetUpsert(true)
	filter := bson.D{{"_id", roomId}}
	update := bson.D{{"$set", bson.D{{"delete", true}}}}

	result, err := workspaceCollection.UpdateOne(database.Ctx, filter, update, opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(result)
}

func GetWorkspaceByUUID(roomUUID string, workspaceUUID string) {

	var room models.Room
	initRoomCollection()
	// find room on UUID

	_ = roomCollection.FindOne(database.Ctx, bson.M{"_id": roomUUID, "workspaces._id": workspaceUUID},
		options.FindOne().SetProjection(bson.M{"workspaces": 1})).Decode(&room)
	fmt.Println(room.Workspaces[0])
	if room.Workspaces == nil {
	}
	fmt.Println(workspaceUUID)
}
