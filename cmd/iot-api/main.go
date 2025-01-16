package main

import (
	"fmt"
	"net/http"
	"time"
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
	mux := http.NewServeMux()

	server := http.Server{
		Addr:              ":8080", // within a container, setting localhost would only enable communication from within the container
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	mux.HandleFunc("/health", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	err := server.ListenAndServe()
	if err != nil {
		fmt.Printf("error in listen and serve: %v", err)
	}

}
