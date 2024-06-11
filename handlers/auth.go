package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"task-management-system/models"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte(os.Getenv("JWT_SECRET_KEY"))

// Register godoc
// @Summary Register a new user
// @Description Register a new user with username, password, role, and email
// @Tags auth
// @Accept  json
// @Produce  json
// @Param user body models.User true "User info"
// @Success 201 {object} models.User
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Router /register [post]
func (db *AppHandler) Register() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Register handler called")
		var user models.User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			log.Println("JSON Decode error: ", err)
			return
		}
		log.Println("User data decoded: ", user)

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Error hashing password: ", err)
			return
		}
		user.Password = string(hashedPassword)
		log.Println("Password hashed:", user.Password)

		_, err = db.DB.Exec("INSERT INTO users (username, password, role, email) VALUES (?, ?, ?, ?)", user.Username, user.Password, user.Role, user.Email)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Database Insert Error: ", err)
			return
		}
		log.Println("User inserted into database: ", user)

		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(user); err != nil {
			log.Println("Error encoding response: ", err)
		}
	})
}

// Login godoc
// @Summary Login a user
// @Description Login a user and return a JWT token
// @Tags auth
// @Accept  json
// @Produce  json
// @Param user body models.User true "User info"
// @Success 200 {object} map[string]string
// @Failure 400 {object} string
// @Failure 401 {object} string
// @Failure 500 {object} string
// @Router /login [post]
func (db *AppHandler) Login() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Login handler called")
		var creds models.User
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			log.Println("JSON Decode error: ", err)
			return
		}
		log.Println("Credentials decoded: ", creds)

		var storedUser models.User
		row := db.DB.QueryRow("SELECT id, username, password, role FROM users WHERE username = ?", creds.Username)
		if err := row.Scan(&storedUser.ID, &storedUser.Username, &storedUser.Password, &storedUser.Role); err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			log.Println("Error fetching user from database: ", err)
			return
		}
		log.Println("User fetched from database: ", storedUser)

		if err := bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(creds.Password)); err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			log.Println("Password comparison error: ", err)
			return
		}
		log.Println("Password matched for user: ", storedUser.Username)

		expirationTime := time.Now().Add(24 * time.Hour)
		claims := &models.Claims{
			Username: storedUser.Username,
			UserID:   storedUser.ID,
			Role:     storedUser.Role,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: expirationTime.Unix(),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(jwtKey)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Error signing token: ", err)
			return
		}
		log.Println("Token generated for user: ", storedUser.Username)

		http.SetCookie(w, &http.Cookie{
			Name:    "token",
			Value:   tokenString,
			Expires: expirationTime,
		})
		log.Println("Cookie set with token for user: ", storedUser.Username)

		if err := json.NewEncoder(w).Encode(map[string]string{"token": tokenString}); err != nil {
			log.Println("Error encoding response: ", err)
		}
	})
}
