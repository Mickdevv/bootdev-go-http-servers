package ApiConfig

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/Mickdevv/bootdev-go-http-servers/api/json_response"
	"github.com/Mickdevv/bootdev-go-http-servers/internal/auth"
	"github.com/Mickdevv/bootdev-go-http-servers/internal/database"
	"github.com/Mickdevv/bootdev-go-http-servers/models"
	"github.com/joho/godotenv"
)

func (cfg *ApiConfig) HandlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}
	type response struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		Token     string `json:"token"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		json_response.RespondWithError(w, http.StatusInternalServerError, "Error unmarshaling login form", err)
	}

	user, err := cfg.DB.GetUserForAuthByEmail(r.Context(), params.Email)
	if err != nil {
		json_response.RespondWithError(w, http.StatusNotFound, "User not found", err)
		return
	}

	password_match, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		json_response.RespondWithError(w, http.StatusInternalServerError, "Error checking password hash", err)
		return
	}
	if !password_match {
		json_response.RespondWithError(w, http.StatusUnauthorized, "Incorrect user credentials", nil)
		return
	}

	if params.ExpiresInSeconds == 0 {
		params.ExpiresInSeconds = 3600
	}

	token, err := auth.MakeJWT(user.ID, cfg.JWTSecret, time.Duration(params.ExpiresInSeconds))
	if err != nil {
		json_response.RespondWithError(w, http.StatusUnauthorized, "Unauthorized: ", err)
		return
	}

	res := response{
		ID:        user.ID.String(),
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Local().String(),
		UpdatedAt: user.UpdatedAt.Local().String(),
		Token:     token,
	}
	json_response.RespondWithJSON(w, http.StatusOK, res)
}

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
		Email    string `json:"email"`
		Password string `json:"password"`
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
	hashed_password, err := auth.HashPassword(params.Password)
	if err != nil {
		json_response.RespondWithError(w, http.StatusInternalServerError, "Error hashing password", err)
		return
	}
	params.Password = hashed_password

	user, err := cfg.DB.CreateUser(r.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: params.Password,
	})
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
