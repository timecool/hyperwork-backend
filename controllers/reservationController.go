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
	"strconv"
	"time"
	"timecool/hyperwork/database"
	"timecool/hyperwork/handler"
	"timecool/hyperwork/models"
)

var reservationCollection *mongo.Collection

func initReservationCollection() {
	database.Ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
	reservationCollection = database.DB.Collection("reservations")
}

func CreateReservation(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Create Reservation")
	w.Header().Add("content-type", "application/json")

	var reservation models.Reservation

	err := json.NewDecoder(r.Body).Decode(&reservation)

	if err != nil {
		handler.HttpErrorResponse(w, http.StatusBadGateway, err.Error())
		return
	}

	// Check if time is in the future
	if time.Now().Unix() > reservation.StartDate {
		handler.HttpErrorResponse(w, http.StatusBadRequest, "Time lies in the past")
		return
	}
	//Check if start time is before end time
	if reservation.EndDate < reservation.StartDate {
		handler.HttpErrorResponse(w, http.StatusBadRequest, "End time is before the start time")
		return
	}

	_, findSome, err := FindReservationBetweenTime(reservation.WorkspaceUUID, reservation.StartDate, reservation.EndDate)

	if err != nil {
		handler.HttpErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	if findSome {
		handler.HttpErrorResponse(w, http.StatusBadRequest, "During this period the workstation has already been reserved")
		return
	}
	user, _ := GetCurrentUser(r)
	reservation.UserUUID = user.UUID
	//Save Names in Object
	roomName, workspaceName := getRoomAndWorkspaceName(reservation.RoomUUID, reservation.WorkspaceUUID)
	if roomName == "" {
		handler.HttpErrorResponse(w, http.StatusNotFound, "Room Name not found")
		return
	}
	if workspaceName == "" {
		handler.HttpErrorResponse(w, http.StatusNotFound, "Workspace Name not found")
		return
	}
	reservation.RoomName = roomName
	reservation.WorkspaceName = workspaceName
	reservation.UUID = uuid.New().String()

	//Save User in Collection User
	result, err := reservationCollection.InsertOne(database.Ctx, reservation)
	if err != nil {
		handler.HttpErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	json.NewEncoder(w).Encode(result)
}

func GetReservationOfUser(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get Reservation Of User")
	w.Header().Add("content-type", "application/json")

	parms := mux.Vars(r)
	userUuid, _ := parms["useruuid"]

	initReservationCollection()
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	startDate, err := strconv.Atoi(r.URL.Query().Get("s"))
	endDate, err := strconv.Atoi(r.URL.Query().Get("e"))

	var filter bson.M
	if int64(endDate) != 0 && int64(startDate) != 0 {
		filter = bson.M{"$and": []interface{}{
			bson.M{
				"$and": []interface{}{
					bson.M{"user_uuid": userUuid},
					getTimeFilter(int64(startDate), int64(endDate)),
				}},
		}}
	} else {
		filter = bson.M{"$and": []interface{}{
			bson.M{"user_uuid": userUuid},
			bson.M{"end_date": bson.M{"$gte": time.Now().Unix()}},
		}}
	}
	var result *mongo.Cursor

	result, err = reservationCollection.Find(database.Ctx, filter, options.Find().SetSort(bson.D{{"start_date", 1}}).SetLimit(int64(limit)))

	if err != nil {
		handler.HttpErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	var reservations []models.Reservation
	if err := result.All(database.Ctx, &reservations); err != nil {
		handler.HttpErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	json.NewEncoder(w).Encode(reservations)

}

func DeleteReservation(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DeleteReservation")
	w.Header().Add("content-type", "application/json")

	params := mux.Vars(r)
	user, err := GetCurrentUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	reservationUUID, _ := params["uuid"]
	initReservationCollection()
	result, err := reservationCollection.DeleteOne(database.Ctx, bson.M{"$and": bson.M{"_id": reservationUUID, "user_uuid": user.UUID}})
	if err != nil {
		handler.HttpErrorResponse(w, http.StatusBadRequest, "Your Reservation is not found")
		return
	}
	json.NewEncoder(w).Encode(result)
}

func GetReservationOfDate(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get Reservation Of Date")
	w.Header().Add("content-type", "application/json")

	params := mux.Vars(r)
	workspaceUUID, _ := params["workspaceuuid"]

	startDate, err := strconv.Atoi(r.URL.Query().Get("s"))
	if err != nil {
		handler.HttpErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	endDate, err := strconv.Atoi(r.URL.Query().Get("e"))
	if err != nil {
		handler.HttpErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	result, _, err := FindReservationBetweenTime(workspaceUUID, int64(startDate), int64(endDate))
	if err != nil {
		handler.HttpErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	json.NewEncoder(w).Encode(result)
}

func FindReservationBetweenTime(workspaceUUID string, start int64, end int64) ([]models.Reservation, bool, error) {
	initReservationCollection()
	result, err := reservationCollection.Find(database.Ctx,
		bson.M{
			"$and": []interface{}{
				bson.M{"workspace_uuid": workspaceUUID},
				getTimeFilter(start, end),
			}},
	)
	if err != nil {
		return nil, false, err
	}
	var reservation []models.Reservation
	if err := result.All(database.Ctx, &reservation); err != nil {
		return nil, false, err
	}

	return reservation, len(reservation) > 0, nil
}

func getTimeFilter(start int64, end int64) bson.M {
	return bson.M{"$or": []interface{}{
		bson.M{"$and": []interface{}{
			bson.M{"start_date": bson.M{"$gte": start}}, // (data.start >= start and data.start <= end)
			bson.M{"start_date": bson.M{"$lte": end}},
		}},
		bson.M{"$and": []interface{}{
			bson.M{"end_date": bson.M{"$lte": start}}, // or (data.end <= start and data.end >= end)
			bson.M{"end_date": bson.M{"$gte": end}},
		}},
		bson.M{"$and": []interface{}{
			bson.M{"start_date": bson.M{"$lte": start}}, //	or (data.start <= start and data.end >= end)
			bson.M{"end_date": bson.M{"$gte": end}},
		}},
	}}
}
