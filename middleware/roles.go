package middleware

import (
	"net/http"
	"timecool/hyperwork/controllers"
	handler2 "timecool/hyperwork/handler"
	"timecool/hyperwork/models"
)

//Member Middleware
//Parm : http.Handler for execution
//return http.Handler
func Member(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//Get Current User
		user, err := controllers.GetCurrentUser(r)
		if err != nil {
			//no user set so no token found
			handler2.HttpErrorResponse(w, http.StatusForbidden, "no token set")
			return
		}
		if user.UserRole == models.RoleMember || user.UserRole == models.RoleAdmin {
			// User has the rights to perform the endpoints
			handler.ServeHTTP(w, r)
			return
		} else {
			// User does not have the rights to perform the endpoints
			handler2.HttpErrorResponse(w, http.StatusUnauthorized, "no right")
			return
		}
	})
}

//Admin Middleware
//Parm : http.Handler for Cookie
//return http.Handler
func Admin(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//Get Current User
		user, err := controllers.GetCurrentUser(r)
		if err != nil {
			//no user set so no token found
			handler2.HttpErrorResponse(w, http.StatusForbidden, "no token set")
			return
		}
		if user.UserRole == models.RoleAdmin {
			// User has the rights to perform the endpoints
			handler.ServeHTTP(w, r)
			return
		} else {
			// User does not have the rights to perform the endpoints
			handler2.HttpErrorResponse(w, http.StatusUnauthorized, "no right")
			return
		}
	})
}
