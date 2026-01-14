package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	_ "github.com/lib/pq"

	ApiConfig "github.com/Mickdevv/bootdev-go-http-servers/api/apiConfig"
	"github.com/Mickdevv/bootdev-go-http-servers/api/chirp"
	"github.com/Mickdevv/bootdev-go-http-servers/api/readiness"
	"github.com/Mickdevv/bootdev-go-http-servers/internal/database"
	"github.com/joho/godotenv"
)

func main() {

	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Error connecting to the database: %s", err)
	}
	dbQueries := database.New(dbConn)

	apiCfg := ApiConfig.ApiConfig{
		FileServerHits: atomic.Int32{},
		DB:             dbQueries,
		JWTSecret:      os.Getenv("JWT_SECRET"),
	}

	mux := http.NewServeMux()

	fsHandler := apiCfg.MiddlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir("./app"))))
	mux.Handle("/app/", fsHandler)

	mux.HandleFunc("GET /admin/metrics", apiCfg.HandlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.HandlerResetUsers)

	mux.HandleFunc("GET /api/healthz", readiness.HandlerReadiness)

	mux.HandleFunc("POST /api/validate_chirp", chirp.HandlerValidateChirp)
	mux.HandleFunc("POST /api/chirps", apiCfg.HandlerCreateChirp)
	mux.HandleFunc("GET /api/chirps", apiCfg.HandlerGetAllChirps)
	mux.HandleFunc("GET /api/chirps/{id}", apiCfg.HandlerGetChirp)

	mux.HandleFunc("POST /api/users", apiCfg.HandlerCreateUser)
	mux.HandleFunc("POST /api/login", apiCfg.HandlerLogin)
	mux.HandleFunc("POST /admin/reset_users", apiCfg.HandlerResetUsers)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	fmt.Println("Server listening on http://localhost:8080")
	log.Fatal(server.ListenAndServe())
}
