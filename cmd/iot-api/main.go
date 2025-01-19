package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/iferdel/sensor-data-streaming-server/internal/pubsub"
	"github.com/iferdel/sensor-data-streaming-server/internal/routing"
	"github.com/iferdel/sensor-data-streaming-server/internal/storage"
	amqp "github.com/rabbitmq/amqp091-go"
)

var apiSettings struct {
	secret string
	dbConn string
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
	// router.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)

	// create api key (further store in database)
	// cli would apply for registration
	// depending on username (for this case) the api key would authorize read-only or all
	// so user registers -> server creates api key and save it in another column from the user table in that specific user's row/record
	// api responds with api key (with https)
	// user store its key locally (it could be done through the cli tool, which could save the api key automatically and refer later on into a dotfile) -- it may also be saved within the cli tool??
	// the user should not share this key
	// anytime the cli tool makes a requests the tool requests includes the api key in the http headers, particularly in "Authorization: Bearer <API_KEY>" header
	// the api key is encrypted (avoid sha-256 bc is fast)
	// the api verifies the api key in every request
	// it extracts the key from the header (or anywhere it is) -> it validates the key with the database -> authorize it or unauthorized (401)
	// role based access control (RBAC)
	// set expiration key (and way to renew the key)
	// database should

	// api endpoints
	// router.HandleFunc("POST /v1/api/register", apiCfg.registerHandler)
	// router.HandleFunc("POST /v1/api/regenerate-key", apiCfg.regenerateKeyHandler)
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
	rabbitConn *amqp.Connection
}

func NewApiConfig() (*apiConfig, error) {
	conn, err := amqp.Dial(routing.RabbitConnString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	return &apiConfig{
		rabbitConn: conn,
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

func (cfg *apiConfig) middelwareLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Printf("%v: %v", req.Method, req.URL.Path)
		next.ServeHTTP(w, req)
	})
}
