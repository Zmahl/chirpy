package main

import (
	"net/http"
	"sort"
	"strconv"
)

func (cfg *apiConfig) getChirps(w http.ResponseWriter, r *http.Request) {
	dbChirps, err := cfg.DB.GetChirps()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirps")
		return
	}

	chirps := []Chirp{}
	authorId := r.URL.Query().Get("author_id")

	if authorId != "" {
		authNumId, err := strconv.Atoi(authorId)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Could not retrieve chirps from that author")
			return
		}
		for _, dbChirp := range dbChirps {
			if dbChirp.AuthorID == authNumId {
				chirps = append(chirps, Chirp{
					ID:   dbChirp.ID,
					Body: dbChirp.Body,
				})
			}
		}
	} else {
		for _, dbChirp := range dbChirps {
			chirps = append(chirps, Chirp{
				ID:   dbChirp.ID,
				Body: dbChirp.Body,
			})
		}
	}

	sort.Slice(chirps, func(i, j int) bool {
		return chirps[i].ID < chirps[j].ID
	})
	sortType := r.URL.Query().Get("sort")
	if sortType == "desc" {
		for i, j := 0, len(chirps)-1; i < j; i, j = i+1, j-1 {
			chirps[i], chirps[j] = chirps[j], chirps[i]
		}
	}

	respondWithJSON(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) getSingleChirp(w http.ResponseWriter, r *http.Request) {
	dbChirps, err := cfg.DB.GetChirps()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirp")
	}

	desiredChirp := Chirp{}
	desiredId, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Id is not a number")
	}
	if desiredId > len(dbChirps) {
		respondWithError(w, http.StatusNotFound, "Chirp Id does not exist yet")
	}
	for _, chirp := range dbChirps {
		if chirp.ID == desiredId {
			desiredChirp.ID = chirp.ID
			desiredChirp.Body = chirp.Body
			break
		}
	}

	respondWithJSON(w, http.StatusOK, desiredChirp)
}
