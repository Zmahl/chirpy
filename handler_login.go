package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Zmahl/chirpy/internal/auth"
)

func (cfg *apiConfig) userLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Expire   int    `json:"expires_in_seconds"`
	}

	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	desiredUser, err := cfg.DB.GetUser(params.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if desiredUser.ID == 0 {
		respondWithError(w, http.StatusUnauthorized, "User with that email does not exist")
		return
	}

	err = auth.CheckPasswordHash(params.Password, string(desiredUser.Password))
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Wrong password")
		return
	}

	defaultExpiration := 60 * 60
	if params.Expire == 0 {
		params.Expire = defaultExpiration
	} else if params.Expire > defaultExpiration {
		params.Expire = defaultExpiration
	}

	token, err := auth.MakeJWT(desiredUser.ID, cfg.SecretString, time.Duration(params.Expire)*time.Second)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create JWT")
	}

	b := make([]byte, 32)
	_, err = rand.Read(b)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not generate refresh token")
	}
	refreshToken := hex.EncodeToString(b)
	err = cfg.DB.RefreshToken(desiredUser.ID, refreshToken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not write refresh token")
	}
	respondWithJSON(w, http.StatusOK, User{
		ID:           desiredUser.ID,
		Email:        desiredUser.Email,
		Token:        token,
		RefreshToken: refreshToken,
		IsRed:        desiredUser.IsRed,
	})
}
