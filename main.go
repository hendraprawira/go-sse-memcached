package main

import (
	"fmt"
	"log"
	"os"

	"alert-map-service/app/db"
	"alert-map-service/app/router"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println(err)
	}

	db, err := db.ConnectDatabase()
	if err != nil {
		log.Fatalln(err)
	}
	db.AutoMigrate()

	port := ":" + os.Getenv("ACTIVE_PORT")
	if err := router.Routes().Run(port); err != nil {
		log.Fatalln(err)
	}
}
