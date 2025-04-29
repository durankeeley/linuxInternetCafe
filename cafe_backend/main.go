package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"cafe_backend/config"
	"cafe_backend/database"
	"cafe_backend/handlers"
	"cafe_backend/middleware"
	"cafe_backend/services"

	"github.com/gorilla/mux"
)

func main() {
	if _, err := os.Stat(database.DatabaseFile); os.IsNotExist(err) {
		log.Println("Database not found, creating new database...")
		database.InitDB()
		password := config.SetupAdminUser()
		log.Printf("Admin user created. Username: administrator, Password: %s\n", password)
	} else {
		database.InitDB()
	}

	go services.WatchSessions() // Watch sessions automatically

	r := mux.NewRouter()

	r.HandleFunc("/api/login", handlers.LoginHandler).Methods("POST")
	r.Handle("/api/computers", middleware.AuthMiddleware(http.HandlerFunc(handlers.AddComputerHandler))).Methods("POST")
	r.Handle("/api/computers", middleware.AuthMiddleware(http.HandlerFunc(handlers.GetComputersHandler))).Methods("GET")
	r.Handle("/api/session/start", middleware.AuthMiddleware(http.HandlerFunc(handlers.StartSession))).Methods("POST")
	r.Handle("/api/session/end", middleware.AuthMiddleware(http.HandlerFunc(handlers.EndSession))).Methods("POST")
	r.Handle("/api/session/unlock", middleware.AuthMiddleware(http.HandlerFunc(handlers.UnlockSession))).Methods("POST")
	r.Handle("/api/session/extend", middleware.AuthMiddleware(http.HandlerFunc(handlers.ExtendSession))).Methods("POST")
	r.Handle("/api/session/notify", middleware.AuthMiddleware(http.HandlerFunc(handlers.NotifySession))).Methods("POST")

	srv := &http.Server{
		Handler:      r,
		Addr:         ":8081",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Println("Cafe Backend server started on :8081")
	log.Fatal(srv.ListenAndServe())
}
