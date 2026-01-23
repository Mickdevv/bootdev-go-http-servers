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
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func (cfg *ApiConfig) HandlerPolkaWebhooks(w http.ResponseWriter, r *http.Request) {
	apiKey, err := auth.GetPolkaKey(r.Header)
	if err != nil {
		json_response.RespondWithError(w, http.StatusUnauthorized, "Api key error", err)
		return
	}

	if apiKey != cfg.PolkaApiKey {
		json_response.RespondWithError(w, http.StatusUnauthorized, "Invalid api key", nil)
		return

	}

	defer r.Body.Close()

	// token, err := auth.GetBearerToken(r.Header)
	// if err != nil {
	// 	json_response.RespondWithError(w, http.StatusUnauthorized, "Authorization error: ", err)
	// 	return
	// }
	//
	// userId, err := auth.ValidateJWT(token, cfg.JWTSecret)
	// if err != nil {
	// 	json_response.RespondWithError(w, http.StatusUnauthorized, "User not authenticated", err)
	// 	return
	// }
	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	type response struct {
		User models.User
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		json_response.RespondWithError(w, http.StatusInternalServerError, "Error unmarshaling user email", err)
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(204)
	}

	_, err = cfg.DB.GetUserById(r.Context(), params.Data.UserID)
	if err != nil {
		json_response.RespondWithError(w, http.StatusNotFound, "User not found", err)
		return
	}

	_, err = cfg.DB.UpgradeUserChirpyRed(r.Context(), database.UpgradeUserChirpyRedParams{
		ID:          params.Data.UserID,
		IsChirpyRed: true,
	})
	if err != nil {
		json_response.RespondWithError(w, http.StatusInternalServerError, "Error updating user", err)
		return
	}

	w.WriteHeader(204)
}

func (cfg *ApiConfig) HandlerRevokeToken(w http.ResponseWriter, r *http.Request) {
	type response struct {
	}

	refresh_token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		json_response.RespondWithError(w, http.StatusUnauthorized, "Authorization error: ", err)
		return
	}

	refresh_token_from_database, err := cfg.DB.GetRefreshTokenByTokenId(r.Context(), refresh_token)

	if err != nil {
		json_response.RespondWithError(w, http.StatusInternalServerError, "Error retrieving refresh token from database", err)
		return
	}
	if refresh_token_from_database.ExpiresAt.Unix() > time.Now().Unix() || refresh_token_from_database.RevokedAt.Valid {
		json_response.RespondWithError(w, http.StatusUnauthorized, "Refresh token invalid", nil)
		return
	}

	revoked_token, err := cfg.DB.RevokeRefreshToken(r.Context(), refresh_token_from_database.Token)
	if err != nil {
		json_response.RespondWithError(w, http.StatusInternalServerError, "Error retrieving token from database", err)
		return
	}
	_ = revoked_token

	json_response.RespondWithJSON(w, http.StatusNoContent, response{})
}

func (cfg *ApiConfig) HandlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Token string `json:"token"`
	}

	refresh_token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		json_response.RespondWithError(w, http.StatusUnauthorized, "Authorization error: ", err)
		return
	}

	refresh_token_from_database, err := cfg.DB.GetRefreshTokenByTokenId(r.Context(), refresh_token)

	if err != nil {
		json_response.RespondWithError(w, http.StatusUnauthorized, "Error retrieving refresh token from database", err)
		return
	}

	if refresh_token_from_database.ExpiresAt.Unix() > time.Now().Unix() || refresh_token_from_database.RevokedAt.Time.Unix() >= time.Now().Unix() {
		json_response.RespondWithError(w, http.StatusUnauthorized, "Refresh token invalid", nil)
		return
	}

	token, err := auth.MakeJWT(refresh_token_from_database.UserID, cfg.JWTSecret, time.Duration(3600))
	if err != nil {
		json_response.RespondWithError(w, http.StatusUnauthorized, "Unauthorized: ", err)
		return
	}

	res := response{
		Token: token,
	}
	json_response.RespondWithJSON(w, http.StatusOK, res)
}

func (cfg *ApiConfig) HandlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}
	type response struct {
		ID           string `json:"id"`
		Email        string `json:"email"`
		CreatedAt    string `json:"created_at"`
		UpdatedAt    string `json:"updated_at"`
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
		IsChirpyRed  bool   `json:"is_chirpy_red"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		json_response.RespondWithError(w, http.StatusInternalServerError, "Error unmarshaling login form", err)
		return
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

	refresh_token, err := auth.MakeRefreshToken()
	if err != nil {
		json_response.RespondWithError(w, http.StatusInternalServerError, "Error creating refresh token", err)
		return
	}

	refresh_token_from_database, err := cfg.DB.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{UserID: user.ID, Token: refresh_token})
	if err != nil {
		json_response.RespondWithError(w, http.StatusInternalServerError, "Error adding refresh token to database", err)
		return
	}

	res := response{
		ID:           user.ID.String(),
		Email:        user.Email,
		CreatedAt:    user.CreatedAt.Local().String(),
		UpdatedAt:    user.UpdatedAt.Local().String(),
		Token:        token,
		RefreshToken: refresh_token_from_database.Token,
		IsChirpyRed:  user.IsChirpyRed,
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

func (cfg *ApiConfig) HandlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		json_response.RespondWithError(w, http.StatusUnauthorized, "Authorization error: ", err)
		return
	}

	userId, err := auth.ValidateJWT(token, cfg.JWTSecret)
	if err != nil {
		json_response.RespondWithError(w, http.StatusUnauthorized, "User not authenticated", err)
		return
	}

	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type response struct {
		User models.User
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
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

	user, err := cfg.DB.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:             userId,
		Email:          params.Email,
		HashedPassword: params.Password,
	})
	if err != nil {
		json_response.RespondWithError(w, http.StatusInternalServerError, "Error updating user", err)
		return
	}

	json_response.RespondWithJSON(w, http.StatusOK, models.User{
		ID:          user.ID,
		Email:       user.Email,
		Created_at:  user.CreatedAt,
		Updated_at:  user.UpdatedAt,
		IsChirpyRed: user.IsChirpyRed,
	})

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
