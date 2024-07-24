package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/Zmahl/chirpy/internal/auth"
)

type Chirp struct {
	ID       int    `json:"id"`
	Body     string `json:"body"`
	AuthorID int    `json:"author_id"`
}

func (cfg *apiConfig) postChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Can't post a chirp while not logged in")
		return
	}

	authorId, err := auth.ValidateJWT(tokenString, cfg.SecretString)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Can't post a chirp while not logged in")
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	authNumId, err := strconv.Atoi(authorId)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user id")
		return
	}

	chirp, err := cfg.DB.CreateChirp(authNumId, cleanseBody(params.Body))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
	}
	respondWithJSON(w, 201, Chirp{
		ID:       chirp.ID,
		Body:     chirp.Body,
		AuthorID: authNumId,
	})
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	if code > 499 {
		log.Printf("Responding with 5XX error: %s", msg)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, errorResponse{
		Error: msg,
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(data)
}

func cleanseBody(s string) string {
	badWords := map[string]bool{
		"kerfuffle": true,
		"sharbert":  true,
		"fornax":    true,
	}

	wordsArr := strings.Split(s, " ")
	for i, w := range wordsArr {
		word := strings.ToLower(w)
		_, exists := badWords[word]
		if exists {
			wordsArr[i] = "****"
		}
	}

	return strings.Join(wordsArr, " ")
}
