package microServerMainFiles

import (
	"encoding/json"
	"log"
	"net/http"
)

func SignUp(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
		return
	}
	var request struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	user := UserCredentials{Email: request.Email}
	if err := RegisterUser(user); err != nil {
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
		return
	}
	var user UserCredentials
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if err := AuthenticateUser(user); err != nil {
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	// Generate JWT Token
	token, err := GenerateJWT(user.Email)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Return the token in the response
	log.Printf("Generated token for user %s: %s", user.Email, token)
	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}
