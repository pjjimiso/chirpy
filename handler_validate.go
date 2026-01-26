package main

import (
	"fmt"
	"encoding/json"
	"net/http"
	"io"
	"regexp"
)

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Body string `json:"body"`
	}
	type responseBody struct {
		CleanedBody string `json:"cleaned_body"`
	}

	dat, err := io.ReadAll(r.Body)
	if err != nil { 
		respondWithError(w, 500, "couldn't read request")
		return
	}

	req := requestBody{}
	err = json.Unmarshal(dat, &req)
	if err != nil { 
		respondWithError(w, 500, "couldn't unmarshal parameters")
		return
	}

	origMsg := req.Body
	if len(origMsg) > 140 { 
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	respondWithJSON(w, 200, responseBody{
		CleanedBody: cleanMessage(origMsg),
	})
}

func cleanMessage(msg string) string {
	cleanedMsg := msg
	badWords := []string{"kerfuffle", "sharbert", "fornax"}
	for _, badWord := range badWords {
		pattern := fmt.Sprintf("(?i)%s", badWord)
		re := regexp.MustCompile(pattern)
		cleanedMsg = re.ReplaceAllString(cleanedMsg, "****")
	}
	return cleanedMsg
}

