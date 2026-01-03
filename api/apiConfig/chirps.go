package ApiConfig

import (
	"encoding/json"
	"net/http"

	"github.com/Mickdevv/bootdev-go-http-servers/api/chirp"
	"github.com/Mickdevv/bootdev-go-http-servers/api/json_response"
	"github.com/Mickdevv/bootdev-go-http-servers/internal/database"
	"github.com/Mickdevv/bootdev-go-http-servers/models"
	"github.com/google/uuid"
)

func (cfg *ApiConfig) HandlerGetChirp(w http.ResponseWriter, r *http.Request) {
	var response models.Chirp

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		json_response.RespondWithError(
			w,
			http.StatusBadRequest,
			"Invalid chirp ID",
			err,
		)
		return
	}
	chirp, err := cfg.DB.GetChirp(r.Context(), id)

	if err != nil {
		json_response.RespondWithError(w, http.StatusNotFound, "Error getting chirp", err)
		return
	}
	response = models.Chirp{
		Created_at: chirp.CreatedAt,
		Updated_at: chirp.UpdatedAt,
		Body:       chirp.Body,
		Id:         chirp.ID,
		User_id:    chirp.UserID,
	}
	json_response.RespondWithJSON(w, http.StatusOK, response)
}

func (cfg *ApiConfig) HandlerGetAllChirps(w http.ResponseWriter, r *http.Request) {
	var response []models.Chirp

	chirps, err := cfg.DB.GetAllChirps(r.Context())

	if err != nil {
		json_response.RespondWithError(w, http.StatusInternalServerError, "Error getting chirps", err)
		return
	}
	for _, c := range chirps {
		response = append(response, models.Chirp{
			Created_at: c.CreatedAt,
			Updated_at: c.UpdatedAt,
			Body:       c.Body,
			Id:         c.ID,
			User_id:    c.UserID,
		})
	}
	json_response.RespondWithJSON(w, http.StatusOK, response)
}

func (cfg *ApiConfig) HandlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	type parameters struct {
		Body   string    `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		json_response.RespondWithError(w, http.StatusInternalServerError, "Could not decode json payload", err)
		return
	}

	params.Body = chirp.ReplaceProfanity(params.Body)

	chirp, err := cfg.DB.CreateChirp(r.Context(), database.CreateChirpParams{
		UserID: params.UserId,
		Body:   params.Body,
	})
	if err != nil {
		json_response.RespondWithError(w, http.StatusInternalServerError, "Error creating the chirp", err)
		return
	}

	json_response.RespondWithJSON(w, http.StatusCreated, models.Chirp{
		Id:         chirp.ID,
		User_id:    chirp.UserID,
		Updated_at: chirp.UpdatedAt,
		Created_at: chirp.CreatedAt,
		Body:       chirp.Body,
	})
}
