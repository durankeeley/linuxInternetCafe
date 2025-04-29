package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

const DatabaseFile = "cafe.db"

func InitDB() {
	var err error
	DB, err = sql.Open("sqlite3", DatabaseFile)
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}

	createTables()
}

func createTables() {
	userTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE,
		password TEXT
	);
	`

	computerTable := `
	CREATE TABLE IF NOT EXISTS computers (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		hostname TEXT,
		ip_address TEXT,
		ssh_port INTEGER DEFAULT 22,
		ssh_username TEXT,
		ssh_private_key TEXT,
		current_password TEXT,
		session_expires_at DATETIME,
		assigned TEXT
	);
	`

	sessionTable := `
	CREATE TABLE IF NOT EXISTS sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		computer_id INTEGER,
		start_time DATETIME,
		end_time DATETIME,
		status TEXT,
		FOREIGN KEY (computer_id) REFERENCES computers(id)
	);
	`

	_, err := DB.Exec(userTable)
	if err != nil {
		log.Fatal("Failed to create users table:", err)
	}

	_, err = DB.Exec(computerTable)
	if err != nil {
		log.Fatal("Failed to create computers table:", err)
	}

	_, err = DB.Exec(sessionTable)
	if err != nil {
		log.Fatal("Failed to create sessions table:", err)
	}
}
