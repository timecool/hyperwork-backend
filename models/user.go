package models

type User struct {
	UUID      string `json:"_id" bson:"_id, omitempty"`
	Name      string `json:"name" bson:"name"`
	Email     string `json:"email,omitempty" bson:"email,omitempty"`
	Password  string `json:"password,omitempty" bson:"password,omitempty"`
	UserRole  Role   `json:"role" bson:"role"`
	LastPlace string `json:"last_place" bson:"last_place"`
}

//Enum in Go
type Role string

const (
	RoleAdmin  = "admin"
	RoleMember = "member"
	RoleNone   = "none"
)
