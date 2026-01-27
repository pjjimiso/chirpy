package main

import "net/http"

func (cfg *apiConfig) handlerAdminReset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(http.StatusText(http.StatusForbidden)))
		return
	}
	w.WriteHeader(http.StatusOK)
	cfg.db.TruncateUsers(r.Context())
	cfg.fileserverHits.Store(0)
	w.Write([]byte("Hits reset to 0"))
}

