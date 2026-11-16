package main

import (
	"context"
	"encoding/base64"
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

// shortcutB64 is the base64-encoded PushHomelabStatus.shortcut file.
// Update by running: base64 < tools/status-widget/PushHomelabStatus.shortcut
const shortcutB64 = "YnBsaXN0MDDcAQIDBAUGBwgJCgsMDXBxKXJ3eHl6e4CBXxARV0ZXb3JrZmxvd0FjdGlvbnNfEBdXRldvcmtmbG93Q2xpZW50UmVsZWFzZV8QF1dGV29ya2Zsb3dDbGllbnRWZXJzaW9uXxAjV0ZXb3JrZmxvd0hhc1Nob3J0Y3V0SW5wdXRWYXJpYWJsZXNeV0ZXb3JrZmxvd0ljb25fEBlXRldvcmtmbG93SW1wb3J0UXVlc3Rpb25zXxAhV0ZXb3JrZmxvd0lucHV0Q29udGVudEl0ZW1DbGFzc2VzXxAeV0ZXb3JrZmxvd01pbmltdW1DbGllbnRWZXJzaW9uXxAkV0ZXb3JrZmxvd01pbmltdW1DbGllbnRWZXJzaW9uU3RyaW5nXxAZV0ZXb3JrZmxvd05vSW5wdXRCZWhhdmlvcl8QIldGV29ya2Zsb3dPdXRwdXRDb250ZW50SXRlbUNsYXNzZXNfEA9XRldvcmtmbG93VHlwZXOjDhcg0g8QERJfEBpXRldvcmtmbG93QWN0aW9uSWRlbnRpZmllcl8QGldGV29ya2Zsb3dBY3Rpb25QYXJhbWV0ZXJzXxAjaXMud29ya2Zsb3cuYWN0aW9ucy5nZXRiYXR0ZXJ5bGV2ZWzSExQVFl8QEEN1c3RvbU91dHB1dE5hbWVUVVVJRF1CYXR0ZXJ5IExldmVsXxAkQUM2NzgyNTQtRDA5Qy00RTNDLThCQTItQkZBQjdFNTg5RTA20g8QGBlfECVpcy53b3JrZmxvdy5hY3Rpb25zLmdldG5ldHdvcmtkZXRhaWxz1BMUGhscHR4fXxAPV0ZOZXR3b3JrRGV0YWlsXxAXV0ZOZXR3b3JrRGV0YWlsc05ldHdvcmtZV2lGaSBOYW1lXxAkQTJFRDAzMTgtOTI5Qy00MDVDLTg2NDktRUQ5QTZBRUM4QjQ2XE5ldHdvcmsgTmFtZVRXaUZp0g8QISJfEB9pcy53b3JrZmxvdy5hY3Rpb25zLmRvd25sb2FkdXJs1iMkJSYnKCkqYWJub1tTaG93SGVhZGVyc1pXRkhUVFBCb2R5XldGSFRUUEJvZHlUeXBlXVdGSFRUUEhlYWRlcnNcV0ZIVFRQTWV0aG9kVVdGVVJMCNIrLC1gVVZhbHVlXxATV0ZTZXJpYWxpemF0aW9uVHlwZdEuL18QG1dGRGljdGlvbmFyeUZpZWxkVmFsdWVJdGVtc6MwQFbTMTIzNDU8WldGSXRlbVR5cGVVV0ZLZXlXV0ZWYWx1ZRAA0issNjvSNzg5Ol8QEmF0dGFjaG1lbnRzQnlSYW5nZVZzdHJpbmfQUmlkXxARV0ZUZXh0VG9rZW5TdHJpbmfSKyw9O9I3OD4/0FlpcGhvbmUtMTXTMTIzNEFF0issQjvSNzhDRNBfEA9iYXR0ZXJ5X3BlcmNlbnTSKyxGO9I3OEdV0UhJVnswLCAxfdRKS0xNThUWVF8QD0FnZ3JhbmRpemVtZW50c1pPdXRwdXROYW1lWk91dHB1dFVVSURUVHlwZaFP0lBRUlNfEBJBZ2dyYW5kaXplbWVudFR5cGVfEBFDb2VyY2lvbkl0ZW1DbGFzc1ZDb2VyY2VfEBNXRk51bWJlckNvbnRlbnRJdGVtXEFjdGlvbk91dHB1dGEgC9MxMjM0V1vSKyxYO9I3OFla0F53aWZpX2Nvbm5lY3RlZNIrLFw70jc4XVXRSF7USktMTV8cHVSgXxAWV0ZEaWN0aW9uYXJ5RmllbGRWYWx1ZVRqc29u0issY2DRLmShZdMxMjM0ZmrSKyxnO9I3OGhp0FxDb250ZW50LVR5cGXSKyxrO9I3OGxt0F8QEGFwcGxpY2F0aW9uL2pzb25UUE9TVF8QH2h0dHA6Ly8xOTIuMTY4LjI5LjIwMTo4MDgwL3B1c2hTNS40WDEzMDAuMC4x0nN0dXZfEBlXRldvcmtmbG93SWNvbkdseXBoTnVtYmVyXxAYV0ZXb3JrZmxvd0ljb25TdGFydENvbG9yEeh3Ejhx3v+goBEDhFM5MDDSfH1+f1ROYW1lWlBhcmFtZXRlcnNfECRXRldvcmtmbG93Tm9JbnB1dEJlaGF2aW9yQXNrRm9ySW5wdXTQoKAACAAhADUATwBpAI8AngC6AN4A/wEmAUIBZwF5AX0BggGfAbwB4gHnAfoB/wINAjQCOQJhAmoCfAKWAqACxwLUAtkC3gMAAw0DGQMkAzMDQQNOA1QDVQNaA2ADdgN5A5cDmwOiA60DswO7A70DwgPHA9wD4wPkA+cD+wQABAUEBgQQBBcEHAQhBCIENAQ5BD4EQQRIBFEEYwRuBHkEfgSABIUEmgSuBLUEywTYBNsE4gTnBOwE7QT8BQEFBgUJBRIFEwUsBTEFNgU5BTsFQgVHBUwFTQVaBV8FZAVlBXgFfQWfBaMFrAWxBc0F6AXrBfAF8QXyBfUF+QX+BgMGDgY1BjYGNwAAAAAAAAIBAAAAAAAAAIIAAAAAAAAAAAAAAAAAAAY4"

func handleShortcut(w http.ResponseWriter, _ *http.Request) {
	data, err := base64.StdEncoding.DecodeString(shortcutB64)
	if err != nil {
		http.Error(w, "shortcut unavailable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", `attachment; filename="PushHomelabStatus.shortcut"`)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
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
	mux.HandleFunc("/shortcut", handleShortcut)

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
