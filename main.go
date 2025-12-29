package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"

	ApiConfig "github.com/Mickdevv/bootdev-go-http-servers/api/apiConfig"
)

func main() {

	apiCfg := ApiConfig.ApiConfig{
		FileServerHits: atomic.Int32{},
	}

	mux := http.NewServeMux()

	fsHandler := apiCfg.MiddlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir("./app"))))
	mux.Handle("/app/", fsHandler)

	mux.HandleFunc("GET /admin/metrics", apiCfg.HandlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.HandlerReset)

	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("content-type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	fmt.Println("Server listening on http://localhost:8080")
	log.Fatal(server.ListenAndServe())
}
