package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Zmahl/chirpy/internal/auth"
)

type TokenResponse struct {
	Token string `json:"token"`
}

func (cfg *apiConfig) refreshJWT(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "User does not have a refresh token")
		return
	}

	id, err := cfg.DB.GetRefreshToken(refreshToken)
	if err != nil {
		fmt.Println("this error")
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	tokenString, err := auth.MakeJWT(id, cfg.SecretString, time.Hour*1)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create new JWT")
		return
	}
	fmt.Println(tokenString)
	respondWithJSON(w, http.StatusOK, TokenResponse{
		Token: tokenString,
	})
}
