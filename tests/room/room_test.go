package room

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
	"timecool/hyperwork/controllers"
	"timecool/hyperwork/database"
	"timecool/hyperwork/models"
)

var testId = ""

var testRoom = []byte(`{
		"name":     "TestRoom",
		"description": "This is a test room"}`)

func TestMain(m *testing.M) {
	database.Connect()
	code := m.Run()

	// remove test room
	database.Ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
	_, _ = database.DB.Collection("rooms").DeleteOne(database.Ctx, bson.M{"_id": testId})
	os.Exit(code)
}

func TestCreateRoom(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/room", bytes.NewBuffer(testRoom))
	w := httptest.NewRecorder()
	controllers.CreateRoom(w, req)

	var m map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &m)

	if m["InsertedID"] == nil {
		t.Errorf("User was not created")
	}
	testId = fmt.Sprintf("%v", m["InsertedID"])
}

func TestGetRooms(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/rooms?page=0&size=1&active=true", nil)
	w := httptest.NewRecorder()
	controllers.GetRooms(w, req)

	var m models.PagingRoom
	json.Unmarshal(w.Body.Bytes(), &m)

	if len(m.Rooms) != 1 {
		t.Errorf("does not return enough users")
	}
	if m.PagingInfo.PageNumber != 0 {
		t.Errorf("return wrong page")
	}
}

func TestRoomByID(t *testing.T) {
	room, isFind, err := controllers.GetRoomById(testId)

	if !isFind {
		t.Errorf("Id not found")
	}
	if err != nil {
		t.Errorf(err.Error())
	}

	if room.UUID != testId || room.Name != "TestRoom" || room.Description != "This is a test room" {
		t.Errorf("Room does not match the test room")
	}
}
