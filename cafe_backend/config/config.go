package config

import (
	"cafe_backend/database"
	"cafe_backend/utils"
	"log"
)

var AdminUsername = "administrator"

func SetupAdminUser() string {
	password := utils.GenerateRandomPassword(12)
	hashedPassword := utils.HashPassword(password)

	_, err := database.DB.Exec(`INSERT INTO users (username, password) VALUES (?, ?)`, AdminUsername, hashedPassword)
	if err != nil {
		log.Fatal("Error creating admin user:", err)
	}

	return password
}
