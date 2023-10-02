package main

import (
	"fmt"
	"jameesjohn.com/uploader-api/config"
	"jameesjohn.com/uploader-api/database"
	"jameesjohn.com/uploader-api/routes"
	"log"
	"net/http"
)

func main() {
	//Setup Env
	config.Load()

	// Setup Database Connection
	database.ConnectDatabase()
	database.Migrate()

	// Setup routes
	router := routes.Router()

	log.Println("Http server running on port", config.Environment.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", config.Environment.Port), router))
}
