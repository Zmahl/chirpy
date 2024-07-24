package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Zmahl/chirpy/internal/auth"
)

func (cfg *apiConfig) upgradeUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			ID int `json:"user_id"`
		} `json:"data"`
	}

	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not identify request")
		return
	}

	polka, _ := auth.GetPolkaKey(r.Header)
	if polka != cfg.PolkaKey {
		fmt.Println(polka)
		respondWithJSON(w, http.StatusUnauthorized, "User cannot upgrade")
		return
	}

	if params.Event != "user.upgraded" {
		respondWithJSON(w, http.StatusNoContent, "")
		return
	}

	userId := params.Data.ID

	err = cfg.DB.UpgradeUser(userId)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Could not upgrade user")
		return
	}

	respondWithJSON(w, http.StatusNoContent, "")
}
