package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/iferdel/sensor-data-streaming-server/internal/pubsub"
	"github.com/iferdel/sensor-data-streaming-server/internal/routing"
	"github.com/iferdel/sensor-data-streaming-server/internal/sensorlogic"
	"github.com/iferdel/sensor-data-streaming-server/internal/storage"
	amqp "github.com/rabbitmq/amqp091-go"
)

var apiSettings struct {
	secret string
	dbConn string
}

type request struct {
	path string
}

func handleRequests(reqs <-chan request) {
	for req := range reqs {
		handleRequest(req)
	}
}

func handleRequest(req request) {
	fmt.Println("handling request from", req.path)
}

const PORT = 8080

func main() {
	router := http.NewServeMux()

	server := http.Server{
		Addr:              ":8080", // within a container, setting localhost would only enable communication from within the container
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	apiCfg, err := NewApiConfig()
	if err != nil {
		log.Fatal(err)
	}
	defer apiCfg.rabbitConn.Close()

	// admin endpoints
	router.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	router.HandleFunc("POST /admin/reset", apiCfg.resetHandler)

	// api endpoints
	router.Handle("GET /v1/api/health", apiCfg.middlewareMetricsInc(http.HandlerFunc(apiCfg.healthHandler)))
	router.HandleFunc("POST /v1/api/validate_command", apiCfg.commandHandler)
	router.HandleFunc("GET /v1/api/sensors", apiCfg.getSensorsHandler)
	router.HandleFunc("GET /v1/api/sensors/{sensorSerialNumber}", apiCfg.getSensorHandler)
	// router.HandleFunc("DELETE /v1/api/sensors/{sensorSerialNumber}", apiCfg.createTargetsHandler)
	router.HandleFunc("GET /v1/api/targets", apiCfg.getTargetsHandler)
	router.HandleFunc("POST /v1/api/targets", apiCfg.createTargetsHandler)
	// router.HandleFunc("DELETE /v1/api/targets/{sensorSerialNumber}", apiCfg.deleteTargetHandler)
	// router.HandleFunc("PUT /api/v1/sensors/{sensorSerialNumber}/target, apiCtf.sensorToTargetHanlder)
	router.HandleFunc("PUT /v1/api/sensors/{sensorSerialNumber}/sleep", apiCfg.sensorSleepHandler)
	router.HandleFunc("PUT /v1/api/sensors/{sensorSerialNumber}/awake", apiCfg.sensorAwakeHandler)
	router.HandleFunc("PUT /v1/api/sensors/{sensorSerialNumber}/change-sample-frequency", apiCfg.sensorChangeSampleFrequencyHandler)

	err = server.ListenAndServe()
	if err != nil {
		fmt.Printf("error in listen and serve: %v", err)
	}
}

type apiConfig struct {
	fileserverHits atomic.Int32
	rabbitConn     *amqp.Connection
}

func NewApiConfig() (*apiConfig, error) {
	conn, err := amqp.Dial(routing.RabbitConnString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	return &apiConfig{
		fileserverHits: atomic.Int32{},
		rabbitConn:     conn,
	}, nil
}

func (cfg *apiConfig) sensorAwakeHandler(w http.ResponseWriter, req *http.Request) {
	sensorSerialNumber := req.PathValue("sensorSerialNumber")

	publishCh, err := cfg.rabbitConn.Channel()
	defer publishCh.Close()
	if err != nil {
		respondWithError(w, 500, "could not create channel to publish sensor's awake command", err)
		return
	}
	err = pubsub.PublishGob(
		publishCh,                // amqp.Channel
		routing.ExchangeTopicIoT, // exchange
		fmt.Sprintf(routing.KeySensorCommandsFormat, sensorSerialNumber)+"."+"awake", // routing key
		routing.SensorCommandMessage{
			SerialNumber: sensorSerialNumber,
			Timestamp:    time.Now(),
			Command:      "awake",
			Params:       nil,
		}, // value
	)
	if err != nil {
		log.Printf("could not publish awake command: %v", err)
	}

}

func (cfg *apiConfig) sensorSleepHandler(w http.ResponseWriter, req *http.Request) {
	sensorSerialNumber := req.PathValue("sensorSerialNumber")

	publishCh, err := cfg.rabbitConn.Channel()
	defer publishCh.Close()
	if err != nil {
		respondWithError(w, 500, "could not create channel to publish sensor's sleep command", err)
		return
	}
	err = pubsub.PublishGob(
		publishCh,                // amqp.Channel
		routing.ExchangeTopicIoT, // exchange
		fmt.Sprintf(routing.KeySensorCommandsFormat, sensorSerialNumber)+"."+"sleep", // routing key
		routing.SensorCommandMessage{
			SerialNumber: sensorSerialNumber,
			Timestamp:    time.Now(),
			Command:      "sleep",
			Params:       nil,
		}, // value
	)
	if err != nil {
		log.Printf("could not publish sleep command: %v", err)
	}

}

func (cfg *apiConfig) sensorChangeSampleFrequencyHandler(w http.ResponseWriter, req *http.Request) {
	// TODO: relation with database, how to keep state between sensor current sample frequency and db registered sample frequency.
	// Maybe this last point (saving sample frequency in db) is redundant and useless

	defer req.Body.Close()

	sensorSerialNumber := req.PathValue("sensorSerialNumber")

	decoder := json.NewDecoder(req.Body)
	type parameters struct {
		NewSampleFrequency float64 `json:"new_sample_frequency"`
	}
	params := parameters{}
	decoder.Decode(&params)

	publishCh, err := cfg.rabbitConn.Channel()
	defer publishCh.Close()
	if err != nil {
		respondWithError(w, 500, "could not create channel to publish sensor's new sample frequency:", err)
		return
	}
	err = pubsub.PublishGob(
		publishCh,                // amqp.Channel
		routing.ExchangeTopicIoT, // exchange
		fmt.Sprintf(routing.KeySensorCommandsFormat, sensorSerialNumber)+"."+"change_sample_frequency", // routing key
		routing.SensorCommandMessage{
			SerialNumber: sensorSerialNumber,
			Timestamp:    time.Now(),
			Command:      "changeSampleFrequency",
			Params: map[string]interface{}{
				"sampleFrequency": params.NewSampleFrequency,
			},
		}, // value
	)
	if err != nil {
		log.Printf("could not publish change sample frequency command: %v", err)
	}

}

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

func (cfg *apiConfig) getTargetsHandler(w http.ResponseWriter, req *http.Request) {
	sensors, err := storage.GetTarget()
	if err != nil {
		log.Printf("Could not retrieve targets: %s", err)
		w.WriteHeader(500)
		return
	}
	respondWithJSON(w, 200, sensors)
}

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

func (cfg *apiConfig) commandHandler(w http.ResponseWriter, req *http.Request) {

	type sensorCommand struct {
		Command string                 `json:"command"` // intended for 'sleep' 'awake' 'changeSampleFrequency'
		Params  map[string]interface{} `json:"params"`
	}

	decoder := json.NewDecoder(req.Body)
	command := sensorCommand{}
	err := decoder.Decode(&command)
	if err != nil {
		log.Printf("Error decoding command: %s", err)
		w.WriteHeader(500)
		return
	}

	if _, exists := sensorlogic.ValidCommands[command.Command]; !exists {
		respondWithError(w, 400, "this is not a valid command", nil)
		return
	}
	// validate params

	respondWithJSON(w, 200, "this is a valid command!")
}

func (cfg *apiConfig) healthHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Hits: %v", cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, req *http.Request) {
	_ = cfg.fileserverHits.Swap(0)
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// do 'middleware' stuff
		cfg.fileserverHits.Add(1)
		// call the next handler
		next.ServeHTTP(w, req)
	})
}

func (cfg *apiConfig) middelwareLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Printf("%v: %v", req.Method, req.URL.Path)
		next.ServeHTTP(w, req)
	})
}

func respondWithError(w http.ResponseWriter, code int, msg string, err error) {
	if err != nil {
		log.Println(err)
	}
	if code > 499 {
		log.Println("Responding with 5XX error:", msg)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, errorResponse{Error: msg})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	dat, err := json.Marshal(payload) // payload accepts any type
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Write(dat)
	return
}
