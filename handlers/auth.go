package handlers

import (
	"auth-backend/database"
	"auth-backend/models"
	"auth-backend/utils"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte(os.Getenv("JWT_SECRET"))

type Credentials struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokenResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

func SignupHandler(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Check if username already exists
	var existingUser models.User
	if err := database.DB.Where("username = ?", creds.Username).First(&existingUser).Error; err == nil {
		http.Error(w, "username already exists", http.StatusBadRequest)
		return
	}

	// Check if email already exists
	var existingEmail models.User
	if err := database.DB.Where("email = ?", creds.Email).First(&existingEmail).Error; err == nil {
		http.Error(w, "email already exists", http.StatusBadRequest)
		return
	}

	// Hash password
	hash, _ := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)

	user := models.User{
		Username: creds.Username,
		Email:    creds.Email,
		Password: string(hash),
	}

	// Create user
	if err := database.DB.Create(&user).Error; err != nil {
		http.Error(w, "failed to create user", http.StatusInternalServerError)
		return
	}

	// Create verification token
	token := uuid.NewString()
	verification := models.VerificationToken{
		UserID: user.ID,
		Token:  token,
		Expiry: time.Now().Add(24 * time.Hour),
	}
	database.DB.Create(&verification)

	// Send email
	link := "http://localhost:8080/verify?token=" + token
	_ = utils.SendMail(user.Email, "Verify your account",
		"<p>Click <a href='"+link+"'>here</a> to verify your account.</p>")

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "user created, check email"})
}

func VerifyHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")

	var vt models.VerificationToken
	if err := database.DB.Where("token = ?", token).First(&vt).Error; err != nil {
		http.Error(w, "invalid or expired token", http.StatusBadRequest)
		return
	}

	if time.Now().After(vt.Expiry) {
		http.Error(w, "token expired", http.StatusBadRequest)
		return
	}

	var user models.User
	database.DB.First(&user, vt.UserID)
	user.IsVerified = true
	database.DB.Save(&user)
	database.DB.Delete(&vt)

	w.Write([]byte("✅ Account verified! You can now log in."))
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	json.NewDecoder(r.Body).Decode(&creds)

	var user models.User
	if err := database.DB.Where("username = ?", creds.Username).First(&user).Error; err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if !user.IsVerified {
		http.Error(w, "please verify your email first", http.StatusUnauthorized)
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)) != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, _ := generateToken(user.Username, 15*time.Minute)
	refreshToken, _ := generateToken(user.Username, 24*time.Hour)

	json.NewEncoder(w).Encode(TokenResponse{Token: token, RefreshToken: refreshToken})
}

func generateToken(username string, duration time.Duration) (string, error) {
	claims := &jwt.RegisteredClaims{
		Subject:   username,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}
