package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/HiteshRepo/awesome-tools/cmd/homelab-agent/internal/metrics"
)

func main() {
	port := flag.Int("port", 9100, "Port to listen on")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/metrics", handleMetrics)

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("homelab-agent listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"ok":true}`))
}

func handleMetrics(w http.ResponseWriter, _ *http.Request) {
	snap, err := metrics.Collect()
	if err != nil {
		http.Error(w, fmt.Sprintf("collection error: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(snap); err != nil {
		log.Printf("encode error: %v", err)
	}
}
