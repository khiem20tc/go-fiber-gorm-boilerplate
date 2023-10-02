package gorm

import (
	"fmt"
	"log"

	config "botp-gateway/config"
	model "botp-gateway/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connection() {
	databaseURL := config.Env("DATABASE_URL")

	database, errConnectDb := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})

	if errConnectDb != nil {
		log.Fatalf("Some error occurred when connect to db. Err: %s", errConnectDb)
	}

	fmt.Println("Connected to database successfully!")
	DB = database
}

func Disconnection() {
	database, err := DB.DB()
	if err != nil {
		log.Fatalf("Some error occurred when disconnect to db. Err: %s", err)
	}
	database.Close()
	fmt.Println("Disconnected to database successfully!")
}

func AutoMigration() {
	DB.AutoMigrate(&model.User{})
}
