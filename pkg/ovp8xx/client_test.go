package ovp8xx

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// newFakeDevice starts an XML-RPC server which answers every "get" request
// with the given JSON, or with HTTP 500 if json is empty.
func newFakeDevice(t *testing.T, json string) (*Client, *atomic.Int64) {
	t.Helper()
	var requests atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if json == "" {
			http.Error(w, "not ready", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/xml")
		_, _ = w.Write([]byte(`<?xml version="1.0"?><methodResponse><params><param><value><string>` +
			json + `</string></value></param></params></methodResponse>`))
	}))
	t.Cleanup(srv.Close)
	return NewClient(WithHost(strings.TrimPrefix(srv.URL, "http://"))), &requests
}

func TestIsAvailable(t *testing.T) {
	client, _ := newFakeDevice(t, `{"device":{"diagnostic":{"confInitStages":["device","ports"]}}}`)
	ok, err := client.IsAvailable(2*time.Second, AndPortsAreOnline())
	if !ok || err != nil {
		t.Fatalf("expected the device to be available, got %v, %v", ok, err)
	}
}

func TestIsAvailableStopsPollingAfterTimeout(t *testing.T) {
	client, requests := newFakeDevice(t, "")
	ok, err := client.IsAvailable(300 * time.Millisecond)
	if ok || err == nil {
		t.Fatalf("expected a timeout, got %v, %v", ok, err)
	}
	// The device must not be flooded with requests
	if n := requests.Load(); n > 10 {
		t.Errorf("too many requests while waiting: %d", n)
	}
	// Polling must stop once IsAvailable returned
	time.Sleep(100 * time.Millisecond) // let an in-flight request finish
	before := requests.Load()
	time.Sleep(500 * time.Millisecond)
	if after := requests.Load(); after != before {
		t.Errorf("polling continued after IsAvailable returned: %d -> %d requests", before, after)
	}
}
