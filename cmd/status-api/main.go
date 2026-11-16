package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/HiteshRepo/awesome-tools/cmd/status-api/internal/api"
	"github.com/HiteshRepo/awesome-tools/cmd/status-api/internal/collector"
	"github.com/HiteshRepo/awesome-tools/cmd/status-api/internal/push"
	"github.com/HiteshRepo/awesome-tools/cmd/status-api/internal/registry"
)

//go:embed widget.js
var widgetJS []byte

func handleWidgetJS(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(widgetJS)
}

func main() {
	port := flag.Int("port", 8080, "Port to listen on")
	devicesPath := flag.String("devices", "/etc/status-api/devices.yaml", "Path to devices.yaml")
	pollInterval := flag.Duration("poll-interval", 60*time.Second, "Agent poll interval")
	pollTimeout := flag.Duration("poll-timeout", 5*time.Second, "Agent poll timeout")
	staleAfter := flag.Duration("stale-after", 2*time.Hour, "iOS push stale threshold")
	flag.Parse()

	devices, err := registry.Load(*devicesPath)
	if err != nil {
		log.Fatalf("load devices: %v", err)
	}
	log.Printf("loaded %d devices from %s", len(devices), *devicesPath)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	agentPoll := collector.NewAgentPoller(devices, *pollInterval, *pollTimeout)
	k8sColl := collector.NewK8sCollector(*pollInterval)
	pushStore := push.NewStore(devices, *staleAfter)

	go agentPoll.Start(ctx)
	go k8sColl.Start(ctx)

	h := api.New(devices, agentPoll, k8sColl, pushStore, *staleAfter)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", api.HandleHealth)
	mux.HandleFunc("/status", h.HandleStatus)
	mux.HandleFunc("/push", pushStore.HandlePush)
	mux.HandleFunc("/widget.js", handleWidgetJS)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		shutCtx, shutCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutCancel()
		srv.Shutdown(shutCtx)
	}()

	log.Printf("status-api listening on :%d", *port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server: %v", err)
	}
}
