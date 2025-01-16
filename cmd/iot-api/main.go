package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/iferdel/sensor-data-streaming-server/internal/sensorlogic"
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

	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
	}

	// admin endpoints
	router.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	router.HandleFunc("POST /admin/reset", apiCfg.resetHandler)

	// api endpoints
	router.Handle("GET /api/health", apiCfg.middlewareMetricsInc(http.HandlerFunc(apiCfg.healthHandler)))
	router.HandleFunc("POST /api/validate_command", apiCfg.commandHandler)

	err := server.ListenAndServe()
	if err != nil {
		fmt.Printf("error in listen and serve: %v", err)
	}
}

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) commandHandler(w http.ResponseWriter, req *http.Request) {

	type sensorCommand struct {
		Command string `json:"command"` // intended for 'sleep' 'awake' 'changeSampleFrequency'
	}

	decoder := json.NewDecoder(req.Body)
	command := sensorCommand{}
	err := decoder.Decode(&command)
	if err != nil {
		log.Printf("Error decoding command: %s", err)
		w.WriteHeader(500)
		return
	}

	type commandResponse struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}

	if _, exists := sensorlogic.ValidCommands[command.Command]; !exists {
		log.Printf("Invalid command received: %v", command.Command)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest) // 400
		dat, err := json.Marshal(commandResponse{
			Status:  "invalid",
			Message: "Invalid command received",
		})
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}
		w.Write(dat)
		return
	}

	dat, err := json.Marshal(commandResponse{Status: "valid", Message: "this is a valid command"})
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(dat)
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
