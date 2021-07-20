package routes

import (
	"github.com/gorilla/mux"
	"net/http"
	"timecool/hyperwork/controllers"
	"timecool/hyperwork/middleware"
	"timecool/hyperwork/models"
)

type route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
	Role        models.Role
}

// base URLS
const baseApiPattern = "/api/v1"
const userPattern = baseApiPattern + "/user"
const roomPattern = baseApiPattern + "/room"
const reservationPattern = baseApiPattern + "/reservation"
const workspacePattern = baseApiPattern + "/workspace"

// return all Endpoints
func getRoutes() []route {
	return []route{
		{
			Name:        "Login",
			Method:      http.MethodPost,
			Pattern:     userPattern + "/login",
			HandlerFunc: controllers.Login,
			Role:        models.RoleNone,
		},
		{
			Name:        "Logout",
			Method:      http.MethodDelete,
			Pattern:     userPattern + "/logout",
			HandlerFunc: controllers.Logout,
			Role:        models.RoleMember,
		},
		{
			Name:        "CreateUser",
			Method:      http.MethodPost,
			Pattern:     userPattern,
			HandlerFunc: controllers.CreateUser,
			Role:        models.RoleNone,
		},
		{
			Name:        "UpdateUser",
			Method:      http.MethodPut,
			Pattern:     userPattern + "/{uuid}",
			HandlerFunc: controllers.UpdateUser,
			Role:        models.RoleAdmin,
		},
		{
			Name:        "SetUserRoles",
			Method:      http.MethodPut,
			Pattern:     userPattern + "/{uuid}/role",
			HandlerFunc: controllers.SetRole,
			Role:        models.RoleAdmin,
		},
		{
			Name:        "GetUserByToken",
			Method:      http.MethodGet,
			Pattern:     userPattern,
			HandlerFunc: controllers.GetUserByToken,
			Role:        models.RoleMember,
		},
		{
			Name:        "GetAllUsers",
			Method:      http.MethodGet,
			Pattern:     userPattern + "s",
			HandlerFunc: controllers.GetUsers,
			Role:        models.RoleAdmin,
		},
		{
			Name:        "GetUser",
			Method:      http.MethodGet,
			Pattern:     userPattern + "/{uuid}",
			HandlerFunc: controllers.GetUser,
			Role:        models.RoleAdmin,
		},
		{
			Name:        "DeleteUser",
			Method:      http.MethodDelete,
			Pattern:     userPattern + "/{uuid}",
			HandlerFunc: controllers.DeleteUser,
			Role:        models.RoleAdmin,
		},
		{
			Name:        "CreateRoom",
			Method:      http.MethodPost,
			Pattern:     roomPattern,
			HandlerFunc: controllers.CreateRoom,
			Role:        models.RoleAdmin,
		},
		{
			Name:        "DeleteRoom",
			Method:      http.MethodDelete,
			Pattern:     roomPattern + "/{uuid}",
			HandlerFunc: controllers.DeleteRoom,
			Role:        models.RoleAdmin,
		},
		{
			Name:        "GetRooms",
			Method:      http.MethodGet,
			Pattern:     roomPattern + "s",
			HandlerFunc: controllers.GetRooms,
			Role:        models.RoleMember,
		},
		{
			Name:        "GetRoom",
			Method:      http.MethodGet,
			Pattern:     roomPattern + "/{uuid}",
			HandlerFunc: controllers.GetRoom,
			Role:        models.RoleMember,
		},
		{
			Name:        "UpdateRoomMap",
			Method:      http.MethodPut,
			Pattern:     roomPattern + "/{uuid}",
			HandlerFunc: controllers.UpdateRoom,
			Role:        models.RoleAdmin,
		},
		{
			Name:        "CreateReservation",
			Method:      http.MethodPost,
			Pattern:     reservationPattern,
			HandlerFunc: controllers.CreateReservation,
			Role:        models.RoleMember,
		},
		{
			Name:        "GetReservationOfDate",
			Method:      http.MethodGet,
			Pattern:     reservationPattern + "/{uuid}",
			HandlerFunc: controllers.GetReservationOfDate,
			Role:        models.RoleMember,
		},
		{
			Name:        "GetReservationOfUser",
			Method:      http.MethodGet,
			Pattern:     userPattern + "/reservation/{uuid}",
			HandlerFunc: controllers.GetReservationOfUser,
			Role:        models.RoleMember,
		},
		{
			Name:        "DeleteReservation",
			Method:      http.MethodDelete,
			Pattern:     reservationPattern + "/{uuid}",
			HandlerFunc: controllers.DeleteReservation,
			Role:        models.RoleNone,
		},
		{
			Name:        "GetWorkspace",
			Method:      http.MethodGet,
			Pattern:     workspacePattern + "/{roomuuid}/{workspaceuuid}",
			HandlerFunc: controllers.GetWorkspace,
			Role:        models.RoleMember,
		},
	}
}

// Params router *mux.Router
// Settings for Endpoints
func Setup(router *mux.Router) {
	for _, route := range getRoutes() {
		switch route.Role {
		case models.RoleAdmin:
			router.Handle(route.Pattern, middleware.Admin(route.HandlerFunc)).Methods(route.Method)
		case models.RoleMember:
			router.Handle(route.Pattern, middleware.Member(route.HandlerFunc)).Methods(route.Method)
		case models.RoleNone:
			router.Handle(route.Pattern, route.HandlerFunc).Methods(route.Method)
		}
	}

}
