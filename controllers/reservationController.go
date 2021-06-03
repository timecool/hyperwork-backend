package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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
	w.Header().Add("content-type", "application")

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
	reservation.UUID = uuid.New().String()

	//Save User in Collection User

	result, err := reservationCollection.InsertOne(database.Ctx, reservation)
	if err != nil {
		handler.HttpErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	json.NewEncoder(w).Encode(result)
}

func GetReservationOfDate(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get Reservation Of Date")
	w.Header().Add("content-type", "application")

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
	fmt.Println(workspaceUUID)
	result, err := reservationCollection.Find(database.Ctx,
		bson.M{
			"$and": []interface{}{
				bson.M{"workspace_uuid": workspaceUUID},
				bson.M{"$or": []interface{}{
					bson.M{"$and": []interface{}{
						bson.M{"start_date": bson.M{"$gte": start}}, //(data.start => start and data.start <= end)
						bson.M{"start_date": bson.M{"$lte": end}},
					}},
					bson.M{"$and": []interface{}{
						bson.M{"end_date": bson.M{"$lte": start}}, //	or (data.end => start and data.end <= end)
						bson.M{"end_date": bson.M{"$gte": end}},
					}},
					bson.M{"$and": []interface{}{
						bson.M{"start_date": bson.M{"$lte": start}}, //	or (data.start <= start and data.end >= end)
						bson.M{"end_date": bson.M{"$gte": end}},
					}},
				}},
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
