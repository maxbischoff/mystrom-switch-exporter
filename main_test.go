package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"testing"
)

type switchData struct {
	temperature     float64
	energySinceBoot float64
	relay           bool
	timeSinceBoot   int
}

func TestHandleCollectRequest(t *testing.T) {
	switchData := switchData{
		temperature:     12.5,
		energySinceBoot: 123478.12,
		relay:           false,
		timeSinceBoot:   2781,
	}
	dummy := getDummySwitchServer(switchData)
	defer dummy.Close()

	req := httptest.NewRequest(http.MethodGet, "/collect?hostname="+dummy.URL, nil)
	w := httptest.NewRecorder()
	handleCollectRequest(w, req)
	res := w.Result()
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("could not read data from http response: %v", err)
	}

	relayNumeric := "0"
	if switchData.relay {
		relayNumeric = "1"
	}
	assertMetricValue(t, "mystrom_switch_temperature", strconv.FormatFloat(switchData.temperature, 'f', -1, 64), data)
	assertMetricValue(t, "mystrom_switch_relay_state", relayNumeric, data)
	assertMetricValue(t, "mystrom_switch_energy_since_boot_wattseconds", strconv.FormatFloat(switchData.energySinceBoot, 'f', -1, 64), data)
	assertMetricValue(t, "mystrom_switch_time_since_boot_seconds", strconv.Itoa(switchData.timeSinceBoot), data)

}

func getDummySwitchServer(data switchData) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, `{
      "power": 29.37,
      "Ws": 30.97,
      "relay": %t,
      "temperature": %f,
      "boot_id": "not used",
      "energy_since_boot": %f,
      "time_since_boot": %d 
		}`, data.relay, data.temperature, data.energySinceBoot, data.timeSinceBoot)
	}))
}

func assertMetricValue(t *testing.T, metricName, expected string, gotMetrics []byte) {
	re, err := regexp.Compile(metricName + " ([0-9]+[.]?[0-9]*)")
	if err != nil {
		t.Errorf("could not compile regex: %v", err)
	}
	found := re.FindSubmatch(gotMetrics)
	if len(found) != 2 {
		t.Errorf("expected metric value for %s in %v", metricName, string(gotMetrics))
	}

	if string(found[1]) != expected {
		t.Errorf("expected %s to be %s but was %s", metricName, expected, found[1])
	}
}
