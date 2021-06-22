package middleware

import (
	"net/http"
	"timecool/hyperwork/controllers"
	handler2 "timecool/hyperwork/handler"
	"timecool/hyperwork/models"
)

func Member(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := controllers.GetCurrentUser(r)
		if err != nil {
			handler2.HttpErrorResponse(w, http.StatusForbidden, "no token set")
			return
		}
		if user.UserRole == models.RoleMember || user.UserRole == models.RoleAdmin {
			handler.ServeHTTP(w, r)
			return
		} else {
			handler2.HttpErrorResponse(w, http.StatusUnauthorized, "no right")
			return
		}
	})
}

func Admin(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := controllers.GetCurrentUser(r)
		if err != nil {
			handler2.HttpErrorResponse(w, http.StatusForbidden, "no token set")
			return
		}
		if user.UserRole == models.RoleAdmin {
			handler.ServeHTTP(w, r)
			return
		} else {
			handler2.HttpErrorResponse(w, http.StatusUnauthorized, "no right")
			return
		}
	})
}
