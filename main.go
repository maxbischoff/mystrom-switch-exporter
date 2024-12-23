package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	var listenAddr string
	flag.StringVar(&listenAddr, "listen-addr", ":8000", "address this exporter listens on")
	flag.Parse()

	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/collect", http.HandlerFunc(handleCollectRequest))
	fmt.Print("exporter started")
	http.ListenAndServe(listenAddr, nil)
}

func handleCollectRequest(w http.ResponseWriter, r *http.Request) {
	hostname := r.URL.Query().Get("hostname")
	if hostname == "" {
		http.Error(w, "must provide 'hostname' request parameter", http.StatusBadRequest)
		return
	}

	registry := prometheus.NewRegistry()
	err := collectSwitchMetrics(r.Context(), hostname, registry)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not collect metrics from %s: %s", hostname, err), http.StatusInternalServerError)
		return
	}

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

type SwitchReport struct {
	Relay           bool    `json:"relay"`
	Temperature     float64 `json:"temperature"`
	EnergySinceBoot float64 `json:"energy_since_boot"`
	TimeSinceBoot   int64   `json:"time_since_boot"`
}

func collectSwitchMetrics(ctx context.Context, addr string, registry *prometheus.Registry) error {
	var (
		temperatureGauge = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "mystrom_switch_temperature",
			Help: "Temperature measured by the device",
		})
		relayGauge = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "mystrom_switch_relay_state",
			Help: "State of the relay, 1 is on and 0 is off",
		})
		energySinceBootGauge = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "mystrom_switch_energy_since_boot_wattseconds",
			Help: "Total energy measured since the last power up or restart in watt seconds",
		})
		timeSinceBootGauge = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "mystrom_switch_time_since_boot_seconds",
			Help: "Time since the last power up or restart in seconds",
		})
	)
	registry.MustRegister(temperatureGauge)
	registry.MustRegister(relayGauge)
	registry.MustRegister(energySinceBootGauge)
	registry.MustRegister(timeSinceBootGauge)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/report", addr), nil)
	if err != nil {
		return fmt.Errorf("could not build get request for call: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("could not get report: %w", err)
	}
	defer resp.Body.Close()

	report := SwitchReport{}
	err = json.NewDecoder(resp.Body).Decode(&report)
	if err != nil {
		return fmt.Errorf("could not decode report: %w", err)
	}

	temperatureGauge.Set(report.Temperature)
	if report.Relay {
		relayGauge.Set(1)
	}
	energySinceBootGauge.Set(report.EnergySinceBoot)
	timeSinceBootGauge.Set(float64(report.TimeSinceBoot))

	return nil
}
