package models

type Room struct {
	UUID          string          `json:"uuid" bson:"_id, omitempty"`
	Name          string          `json:"name,omitempty" bson:"name,omitempty"`
	Description   string          `json:"description" bson:"description"`
	Workspaces    []Workspace     `json:"workspaces" bson:"workspaces"`
	Specification []Specification `json:"specification" bson:"specification"`
}
