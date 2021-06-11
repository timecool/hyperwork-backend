package models

type Workspace struct {
	UUID        string  `json:"uuid" bson:"_id, omitempty"`
	Name        string  `json:"name,omitempty" bson:"name,omitempty"`
	Description string  `json:"description" bson:"description"`
	PositionX   float64 `json:"positionX" bson:"position_x, omitempty"`
	PositionY   float64 `json:"positionY" bson:"position_y, omitempty"`
	Rotate      float64 `json:"rotate" bson:"rotate, omitempty"`
}
