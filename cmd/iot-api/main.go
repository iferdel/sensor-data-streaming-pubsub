package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/iferdel/sensor-data-streaming-server/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

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

const PORT = "8080"

func main() {
	router := http.NewServeMux()

	server := http.Server{
		Addr:              ":" + PORT, // within a container, setting localhost would only enable communication from within the container
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	apiCfg, err := NewApiConfig()
	if err != nil {
		log.Fatal(err)
	}
	defer apiCfg.rabbitConn.Close()

	// api endpoints
	// router.HandleFunc("POST /v1/api/register", apiCfg.handlerAccountRegister)
	// router.HandleFunc("POST /v1/api/regenerate-key", apiCfg.handlerAccountRegenerateKey)
	router.HandleFunc("GET /v1/api/sensors", apiCfg.handlerSensorsRetrieve)
	router.HandleFunc("GET /v1/api/sensors/{sensorSerialNumber}", apiCfg.handlerSensorsGet)
	// router.HandleFunc("DELETE /v1/api/sensors/{sensorSerialNumber}", apiCfg.handlerTargetsCreate)
	router.HandleFunc("GET /v1/api/targets", apiCfg.handlerTargetsGet)
	router.HandleFunc("POST /v1/api/targets", apiCfg.handlerTargetsCreate)
	// router.HandleFunc("DELETE /v1/api/targets/{sensorSerialNumber}", apiCfg.handlerTargetsDelete)
	// router.HandleFunc("PUT /api/v1/sensors/{sensorSerialNumber}/target, apiCtf.handlerSensorsToTarget)
	router.HandleFunc("PUT /v1/api/sensors/{sensorSerialNumber}/sleep", apiCfg.handlerSensorsSleep)
	router.HandleFunc("PUT /v1/api/sensors/{sensorSerialNumber}/awake", apiCfg.handlerSensorsAwake)
	router.HandleFunc("PUT /v1/api/sensors/{sensorSerialNumber}/change-sample-frequency", apiCfg.handlerSensorsChangeSampleFrequency)

	err = server.ListenAndServe()
	if err != nil {
		fmt.Printf("error in listen and serve: %v", err)
	}
}

func (cfg *apiConfig) middelwareLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Printf("%v: %v", req.Method, req.URL.Path)
		next.ServeHTTP(w, req)
	})
}
