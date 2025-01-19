package main

import (
	"log"
	"net/http"

	"github.com/iferdel/sensor-data-streaming-server/internal/storage"
)

func (cfg *apiConfig) getSensorsHandler(w http.ResponseWriter, req *http.Request) {

	sensors, err := storage.GetSensor()
	if err != nil {
		log.Printf("Could not retrieve sensors: %s", err)
		w.WriteHeader(500)
		return
	}
	respondWithJSON(w, 200, sensors)
}

func (cfg *apiConfig) getSensorHandler(w http.ResponseWriter, req *http.Request) {

	sensorSerialNumber := req.PathValue("sensorSerialNumber")
	sensor, err := storage.GetSensorBySerialNumber(sensorSerialNumber)

	if err != nil {
		log.Printf("Could not retrieve sensor %v: %s", sensorSerialNumber, err)
		w.WriteHeader(500)
		return
	}
	respondWithJSON(w, 200, sensor)
}

