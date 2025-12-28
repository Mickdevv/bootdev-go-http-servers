package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

func main() {

	apiCfg := apiConfig{
		fileServerHits: atomic.Int32{},
	}

	mux := http.NewServeMux()

	fsHandleHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir("./app"))))
	mux.Handle("/app/", fsHandleHandler)

	mux.HandleFunc("GET /metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /reset", apiCfg.handlerReset)

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, req *http.Request) {
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
