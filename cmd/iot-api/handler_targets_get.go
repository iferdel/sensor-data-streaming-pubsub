package main

import (
	"log"
	"net/http"
)

func (cfg *apiConfig) handlerTargetsGet(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	sensors, err := cfg.db.GetTarget(ctx)
	if err != nil {
		log.Printf("Could not retrieve targets: %s", err)
		w.WriteHeader(500)
		return
	}
	respondWithJSON(w, 200, sensors)
}
