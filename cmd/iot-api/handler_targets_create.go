package main

import (
	"encoding/json"
	"net/http"

	"github.com/iferdel/sensor-data-streaming-server/internal/storage"
)

func (cfg *apiConfig) createTargetsHandler(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	decoder := json.NewDecoder(req.Body)
	params := storage.TargetRecord{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Error decoding create target parameters", err)
		return
	}

	err = storage.WriteTarget(params)
	if err != nil {
		respondWithError(w, 500, "Could not create new target", err)
		return
	}

	respondWithJSON(w, 201, "Target created!")
}
