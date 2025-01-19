package main

import (
	"log"
	"net/http"

	"github.com/iferdel/sensor-data-streaming-server/internal/storage"
)

func (cfg *apiConfig) getTargetsHandler(w http.ResponseWriter, req *http.Request) {
	sensors, err := storage.GetTarget()
	if err != nil {
		log.Printf("Could not retrieve targets: %s", err)
		w.WriteHeader(500)
		return
	}
	respondWithJSON(w, 200, sensors)
}
