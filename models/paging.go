package models

type PagingUser struct {
	PagingInfo struct {
		PageNumber int64 `json:"pageNumber"`
		Steps      int64 `json:"steps"`
		TotalItems int64 `json:"totalItems"`
	} `json:"pagingInfo"`
	Users []User `json:"users"`
}
