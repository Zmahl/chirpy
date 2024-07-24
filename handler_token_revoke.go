package main

import (
	"net/http"

	"github.com/Zmahl/chirpy/internal/auth"
)

func (cfg *apiConfig) revokeJWT(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "User does not have a refresh token")
		return
	}

	err = cfg.DB.RevokeToken(refreshToken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not delete given refresh token")
		return
	}

	respondWithJSON(w, http.StatusNoContent, "")
}
