package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Zmahl/chirpy/internal/auth"
)

type User struct {
	ID           int    `json:"id"`
	Email        string `json:"email"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	IsRed        bool   `json:"is_chirpy_red"`
}

type UserLogin struct {
	ID     int    `json:"id"`
	Email  string `json:"email"`
	Expire int    `json:"expires_in_seconds,omitempty"`
}

func (cfg *apiConfig) updateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT")
		return
	}
	subject, err := auth.ValidateJWT(token, cfg.SecretString)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't hash password")
	}

	numId, err := strconv.Atoi(subject)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't parse user ID")
		return
	}

	err = cfg.DB.UpdateUser(numId, params.Email, hashedPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create user")
		return
	}
	respondWithJSON(w, http.StatusOK, User{
		ID:    numId,
		Email: params.Email,
	})
}
