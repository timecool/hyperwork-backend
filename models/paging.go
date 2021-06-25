package models

type PagingInfo struct {
	PageNumber int64 `json:"pageNumber"`
	Steps      int64 `json:"steps"`
	TotalItems int64 `json:"totalItems"`
}
type PagingUser struct {
	PagingInfo PagingInfo `json:"pagingInfo"`
	Users      []User     `json:"users"`
}

type PagingRoom struct {
	PagingInfo PagingInfo `json:"pagingInfo"`
	Rooms      []Room     `json:"rooms"`
}
