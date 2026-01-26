package main

import "net/http"

func (cfg *apiConfig) handlerResetRequestCount(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)
	w.Write([]byte("Hits reset to 0"))
}

