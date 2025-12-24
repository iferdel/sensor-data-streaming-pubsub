package main

import "net/http"

func (cfg *apiConfig) HandlerReadiness(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")

	err := cfg.db.Ping(ctx)
	if err != nil {
		http.Error(w, "Database pool not ready", http.StatusServiceUnavailable)
		return
	}

	if cfg.rabbitConn.IsClosed() {
		http.Error(w, "Messaging system not ready", http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}
