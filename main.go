package main

import (
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"timecool/hyperwork/database"
	"timecool/hyperwork/routes"
	"timecool/hyperwork/util"
)

func main() {
	//Database configuration and connection
	database.Connect()
	//Create new Router for Endpointer
	r := mux.NewRouter().StrictSlash(true)
	// Create all Entpoints
	routes.Setup(r)

	fmt.Println("Setup set")
	log.Fatal(http.ListenAndServe(util.GetEnvVariable("API_PORT"),
		//Settings Header Cors
		handlers.CORS(
			handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}),
			handlers.AllowedMethods([]string{"GET", "POST", "DELETE", "PUT"}),
			handlers.AllowCredentials(),
			handlers.AllowedOrigins([]string{util.GetEnvVariable("SERVER_URL"), util.GetEnvVariable("MOBILE_URL")}),
		)(r)))
}
