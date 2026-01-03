package ApiConfig

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/Mickdevv/bootdev-go-http-servers/api/json_response"
	"github.com/Mickdevv/bootdev-go-http-servers/models"
	"github.com/joho/godotenv"
)

func (cfg *ApiConfig) HandlerResetUsers(w http.ResponseWriter, r *http.Request) {
	godotenv.Load()

	ENV := os.Getenv("PLATFORM")
	if ENV != "DEV" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Error, endpoint can only be called in DEV environment"))
		return
	}

	err := cfg.DB.DeleteAllUsers(r.Context())
	if err != nil {
		json_response.RespondWithError(w, http.StatusInternalServerError, "Error deleting data", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (cfg *ApiConfig) HandlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
	}
	type response struct {
		User models.User
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		json_response.RespondWithError(w, http.StatusInternalServerError, "Error unmarshaling user email", err)
		return
	}

	user, err := cfg.DB.CreateUser(r.Context(), params.Email)
	if err != nil {
		json_response.RespondWithError(w, http.StatusInternalServerError, "Error creating user", err)
		return
	}

	json_response.RespondWithJSON(w, http.StatusCreated, models.User{
		ID:         user.ID,
		Email:      user.Email,
		Created_at: user.CreatedAt,
		Updated_at: user.UpdatedAt,
	})

}
