package database

import (
	"log"

	"auth-backend/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	db, err := gorm.Open(sqlite.Open("auth.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to database: ", err)
	}

	db.AutoMigrate(&models.User{})

	DB = db
	log.Println("Database connection established")
}
