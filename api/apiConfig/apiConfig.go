package ApiConfig

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/Mickdevv/bootdev-go-http-servers/api/json_response"
	"github.com/Mickdevv/bootdev-go-http-servers/internal/database"
	"github.com/Mickdevv/bootdev-go-http-servers/models"
)

type ApiConfig struct {
	FileServerHits atomic.Int32
	DB             *database.Queries
}

func (cfg *ApiConfig) HandlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		json_response.RespondWithError(w, http.StatusInternalServerError, "Error unmarshaling user email", err)
	}

	user, err := cfg.DB.CreateUser(r.Context(), params.Email)
	if err != nil {
		json_response.RespondWithError(w, http.StatusInternalServerError, "Error creating user", err)
	}

	json_response.RespondWithJSON(w, http.StatusCreated, user)

}
func (cfg *ApiConfig) HandlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	html := fmt.Sprintf(fmt.Sprintf(`
<html>

<body>
	<h1>Welcome, Chirpy Admin</h1>
	<p>Chirpy has been visited %d times!</p>
</body>

</html>
	`, cfg.FileServerHits.Load()))

	w.Write([]byte(html))

}

func (cfg *ApiConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.FileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *ApiConfig) HandlerReset(w http.ResponseWriter, r *http.Request) {
	cfg.FileServerHits.Store(0)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0"))
}
