package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/iferdel/sensor-data-streaming-server/internal/routing"
	"github.com/iferdel/sensor-data-streaming-server/internal/storage"
	amqp "github.com/rabbitmq/amqp091-go"
)

type apiConfig struct {
	ctx        context.Context
	rabbitConn *amqp.Connection
	db         *storage.DB
}

func NewApiConfig() (*apiConfig, error) {
	ctx := context.Background()

	conn, err := amqp.Dial(routing.RabbitConnString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	db, err := storage.NewDBPool(storage.PostgresConnString)

	return &apiConfig{
		ctx:        ctx,
		rabbitConn: conn,
		db:         db,
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
	defer apiCfg.db.Close()

	// api endpoints
	// router.HandleFunc("POST /api/v1/register", apiCfg.handlerAccountRegister)
	// router.HandleFunc("POST /api/v1/regenerate-key", apiCfg.handlerAccountRegenerateKey)
	router.HandleFunc("GET /api/v1/sensors", apiCfg.handlerSensorsRetrieve)
	router.HandleFunc("GET /api/v1/sensors/{sensorSerialNumber}", apiCfg.handlerSensorsGet)
	// router.HandleFunc("DELETE /api/v1/sensors/{sensorSerialNumber}", apiCfg.handlerTargetsCreate)
	router.HandleFunc("GET /api/v1/targets", apiCfg.handlerTargetsGet)
	router.HandleFunc("POST /api/v1/targets", apiCfg.handlerTargetsCreate)
	// router.HandleFunc("DELETE /api/v1/targets/{sensorSerialNumber}", apiCfg.handlerTargetsDelete)
	// router.HandleFunc("PUT /api/v1/sensors/{sensorSerialNumber}/target, apiCtf.handlerSensorsToTarget)
	router.HandleFunc("PUT /api/v1/sensors/{sensorSerialNumber}/sleep", apiCfg.handlerSensorsSleep)
	router.HandleFunc("PUT /api/v1/sensors/{sensorSerialNumber}/awake", apiCfg.handlerSensorsAwake)
	router.HandleFunc("PUT /api/v1/sensors/{sensorSerialNumber}/change-sample-frequency", apiCfg.handlerSensorsChangeSampleFrequency)

	err = server.ListenAndServe()
	if err != nil {
		fmt.Printf("error in listen and serve: %v", err)
	}
}

// func (cfg *apiConfig) middelwareLog(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
// 		log.Printf("%v: %v", req.Method, req.URL.Path)
// 		next.ServeHTTP(w, req)
// 	})
// }
