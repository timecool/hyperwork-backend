package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
	"timecool/hyperwork/database"
	"timecool/hyperwork/handler"
	"timecool/hyperwork/models"
)

func GetWorkspace(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get Workspace")
	w.Header().Add("content-type", "application/json")
	params := mux.Vars(r)
	roomUUID, _ := params["roomuuid"]
	workspaceUUID, _ := params["workspaceuuid"]
	fmt.Println(workspaceUUID)
	workspace, err := GetWorkspaceByUUID(roomUUID, workspaceUUID)
	if err != nil {
		handler.HttpErrorResponse(w, http.StatusNotFound, err.Error())
		return
	}
	json.NewEncoder(w).Encode(workspace)
}

func GetWorkspaceByUUID(roomUUID string, workspaceUUID string) (models.Room, error) {

	var room models.Room
	initRoomCollection()
	// find room on UUID
	err := roomCollection.FindOne(database.Ctx, bson.M{"_id": roomUUID, "workspaces._id": workspaceUUID},
		options.FindOne().SetProjection(bson.M{"workspaces.$": 1, "name": 1})).Decode(&room)
	if err != nil {
		return room, err
	}
	fmt.Println(room)
	return room, err
}
