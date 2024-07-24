package main

import (
	"net/http"
	"strconv"

	"github.com/Zmahl/chirpy/internal/auth"
)

func (cfg *apiConfig) deleteChirp(w http.ResponseWriter, r *http.Request) {
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	authId, err := auth.ValidateJWT(tokenString, cfg.SecretString)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	chirps, err := cfg.DB.GetChirps()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get chirps")
		return
	}

	authNumId, err := strconv.Atoi(authId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not validate author id")
		return
	}

	chirpNumId, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Chirp id is not a number")
		return
	}

	for _, chirp := range chirps {
		if chirp.ID == chirpNumId {
			if chirp.AuthorID == authNumId {
				cfg.DB.DeleteChirp(chirpNumId)
			} else {
				respondWithError(w, http.StatusForbidden, "User cannot delete this chirp")
				return
			}
		}
	}

	respondWithJSON(w, http.StatusNoContent, "")
}
