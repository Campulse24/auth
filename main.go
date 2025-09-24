package main

import (
	"auth-backend/database"
	"auth-backend/handlers"
	"auth-backend/middleware"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	database.Connect()

	r := chi.NewRouter()

	// Public routes
	r.Post("/signup", handlers.SignupHandler)
	r.Post("/login", handlers.LoginHandler)
	r.Get("/verify", handlers.VerifyHandler)

	// Protected
	r.Get("/profile", middleware.AuthMiddleware(ProfileHandler))

	log.Println("Server running on :8080")
	http.ListenAndServe(":8080", r)
}

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Welcome to your profile! ✅"))
}
