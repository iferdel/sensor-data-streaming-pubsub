package main

import (
	"log"
	"net/http"
)

func (cfg *apiConfig) handlerSensorsGet(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	sensorSerialNumber := req.PathValue("sensorSerialNumber")
	sensor, err := cfg.db.GetSensorBySerialNumber(ctx, sensorSerialNumber)

	if err != nil {
		log.Printf("Could not retrieve sensor %v: %s", sensorSerialNumber, err)
		w.WriteHeader(500)
		return
	}
	respondWithJSON(w, 200, sensor)
}

func (cfg *apiConfig) handlerSensorsRetrieve(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	sensors, err := cfg.db.GetSensor(ctx)
	if err != nil {
		log.Printf("Could not retrieve sensors: %s", err)
		w.WriteHeader(500)
		return
	}
	respondWithJSON(w, 200, sensors)
}
