package main


import (
	"net/http"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/pjjimiso/chirpy/internal/auth"
)


func (cfg *apiConfig) handlerPolkaWebhooks(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Event string `json:"event"`
		Data struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil { 
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode json parameters", err)
		return
	}

	key, err := auth.GetAPIKey(r.Header)
	if key != cfg.polkaApiKey || err != nil { 
		respondWithError(w, http.StatusUnauthorized, "Invalid API key", err)
		return
	}

	switch params.Event {
	case "user.upgraded":
		err = cfg.db.UpdateUserAddChirpyRed(r.Context(), params.Data.UserID)
		if err != nil {
			respondWithError(w, http.StatusNotFound, "Couldn't upgrade user", err)
			return
		}
	default:
		respondWithError(w, http.StatusNoContent, "That event isn't supported", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
