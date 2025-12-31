package chirp

import (
	"encoding/json"
	"net/http"
	"strings"

	json_response "github.com/Mickdevv/bootdev-go-http-servers/api/json_response"
)

func HandlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	type validateChirpResponse struct {
		Valid bool   `json:"valid"`
		Body  string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	chirp := parameters{}
	err := decoder.Decode(&chirp)
	if err != nil {
		json_response.RespondWithError(w, http.StatusInternalServerError, "Error unmarshaling chirp", err)
		return
	}

	if len(chirp.Body) > 140 {
		json_response.RespondWithError(w, http.StatusBadRequest, "Chirp body too long", nil)
		return
	}

	chirp.Body = replaceProfanity(chirp.Body)
	respBody := validateChirpResponse{Valid: true, Body: replaceProfanity(chirp.Body)}
	json_response.RespondWithJSON(w, http.StatusOK, respBody)

}

func replaceProfanity(s string) string {
	profanities := []string{"kerfuffle", "sharbert", "fornax"}

	returnString := s
	for _, profanity := range profanities {
		for strings.Contains(strings.ToLower(returnString), profanity) {
			occurrenceIndex := strings.Index(strings.ToLower(returnString), profanity)
			returnString = returnString[:occurrenceIndex] + "****" + returnString[occurrenceIndex+len(profanity):]
		}
	}
	return returnString
}
