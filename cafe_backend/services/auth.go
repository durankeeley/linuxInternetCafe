package services

import (
	"cafe_backend/database"
	"cafe_backend/utils"
)

func Authenticate(username, password string) (bool, error) {
	var hashedPassword string
	err := database.DB.QueryRow(`SELECT password FROM users WHERE username = ?`, username).Scan(&hashedPassword)
	if err != nil {
		return false, err
	}

	return utils.CheckPasswordHash(password, hashedPassword), nil
}
