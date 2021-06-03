package models

type Specification struct {
	Name      string  `json:"name,omitempty" bson:"name,omitempty"`
	PositionX float64 `json:"positionX" bson:"position_x, omitempty"`
	PositionY float64 `json:"positionY" bson:"position_y, omitempty"`
	Rotate    float64 `json:"rotate" bson:"rotate, omitempty"`
}
