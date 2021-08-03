package user

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

func TestMain(m *testing.M) {
	fmt.Println("test start")
	database.Connect()
	code := m.Run()

	// remove test user
	database.Ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
	_, _ = database.DB.Collection("users").DeleteOne(database.Ctx, bson.M{"email": "vincenttest@test.gg"})
	os.Exit(code)
}

var testUser = []byte(`{
		"name":     "Vincent",
		"email":    "vincenttest@test.gg",
		"password": "MyPersonalPassword"}`)

func TestCreateUser(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/user", bytes.NewBuffer(testUser))
	w := httptest.NewRecorder()
	controllers.CreateUser(w, req)

	var m map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &m)

	if m["InsertedID"] == nil {
		t.Errorf("User was not created")
	}
}

func TestCreateUserDoppleEmail(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/user", bytes.NewBuffer(testUser))
	w := httptest.NewRecorder()
	controllers.CreateUser(w, req)

	var m map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &m)

	if m["message"] == "Email already exists" {
		t.Logf("Incorrect input was detected")
	}
	if m["InsertedID"] != nil {
		t.Errorf("User could be created twice")
	}
}

func TestGetUsers(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users?page=0&size=1&active=true", nil)
	w := httptest.NewRecorder()
	controllers.GetUsers(w, req)

	var m models.PagingUser
	json.Unmarshal(w.Body.Bytes(), &m)

	if len(m.Users) != 1 {
		t.Errorf("does not return enough users")
	}
	if m.PagingInfo.PageNumber != 0 {
		t.Errorf("return wrong page")
	}
}

func TestGetUserByEmail(t *testing.T) {
	user, isFound, err := controllers.GetUserByEmail("vincenttest@test.gg")

	if !isFound {
		t.Errorf("User not found")
	}
	if user.Name != "Vincent" {
		t.Errorf("Wrong User found")
	}
	if err != nil {
		t.Errorf("Error found")
	}
}
