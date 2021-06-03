package models

type Reservation struct {
	UUID          string `json:"uuid" bson:"_id, omitempty"`
	WorkspaceUUID string `json:"workspace_uuid" bson:"workspace_uuid, omitempty"`
	UserUUID      string `json:"user_uuid" bson:"user_uuid, omitempty"`
	StartDate     int64  `json:"start_date,omitempty" bson:"start_date, omitempty"`
	EndDate       int64  `json:"end_date,omitempty" bson:"end_date, omitempty"`
}
