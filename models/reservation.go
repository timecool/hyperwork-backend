package models

type Reservation struct {
	UUID          string `json:"uuid" bson:"_id, omitempty"`
	WorkspaceUUID string `json:"workspaceUuid" bson:"workspace_uuid, omitempty"`
	UserUUID      string `json:"userUuid" bson:"user_uuid, omitempty"`
	StartDate     int64  `json:"startDate,omitempty" bson:"start_date, omitempty"`
	EndDate       int64  `json:"endDate,omitempty" bson:"end_date, omitempty"`
}
