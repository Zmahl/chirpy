package main

import (
	"encoding/json"
	"net/http"
)

func (cfg *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	existingUser, err := cfg.DB.GetUser(params.Email)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	if existingUser.ID != 0 {
		respondWithError(w, 401, "User  with that email already exists")
		return
	}

	user, err := cfg.DB.CreateUser(params.Email, params.Password)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
	}
	respondWithJSON(w, 201, User{
		user.ID,
		user.Email,
		"",
		"",
		false,
	})
}
