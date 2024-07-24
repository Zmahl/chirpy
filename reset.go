package main

import "net/http"

func (cfg *apiConfig) reset(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	cfg.fileServerHits = 0
}
