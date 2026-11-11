package ovp8xx

import (
	"net"
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

func TestClientURL(t *testing.T) {
	tests := []struct{ host, want string }{
		{"192.168.0.69", "http://192.168.0.69/api/rpc/v1/com.ifm.efector/"},
		{"192.168.0.69:8080", "http://192.168.0.69:8080/api/rpc/v1/com.ifm.efector/"},
		{"ovp.local", "http://ovp.local/api/rpc/v1/com.ifm.efector/"},
		{"::1", "http://[::1]/api/rpc/v1/com.ifm.efector/"},
		{"fe80::1%eth0", "http://[fe80::1%25eth0]/api/rpc/v1/com.ifm.efector/"},
		{"[::1]:8080", "http://[::1]:8080/api/rpc/v1/com.ifm.efector/"},
	}
	for _, tt := range tests {
		c := NewClient(WithHost(tt.host))
		if c.url != tt.want {
			t.Errorf("host %q: got %q, want %q", tt.host, c.url, tt.want)
		}
	}
	d := NewClient(WithHost("::1")).GetDiagnosticClient()
	if want := "http://[::1]/api/rpc/v1/com.ifm.diagnostic/"; d.url != want {
		t.Errorf("diagnostic client: got %q, want %q", d.url, want)
	}
}

func TestIsAvailableIPv6(t *testing.T) {
	l, err := net.Listen("tcp", "[::1]:0")
	if err != nil {
		t.Skip("IPv6 loopback not available:", err)
	}
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml")
		_, _ = w.Write([]byte(`<?xml version="1.0"?><methodResponse><params><param><value><string>` +
			`{"device":{"diagnostic":{"confInitStages":["device"]}}}</string></value></param></params></methodResponse>`))
	}))
	srv.Listener = l
	srv.Start()
	t.Cleanup(srv.Close)
	// The host is given as "[::1]:port", as a user would pass it on the command line
	client := NewClient(WithHost(l.Addr().String()))
	if ok, err := client.IsAvailable(2 * time.Second); !ok || err != nil {
		t.Fatalf("expected the device to be reachable via IPv6, got %v, %v", ok, err)
	}
}

func TestNewClientIgnoresEnvironment(t *testing.T) {
	t.Setenv("OVP8XX_IP", "10.0.0.1")
	if c := NewClient(); c.host != DefaultHost {
		t.Errorf("expected %q, got %q", DefaultHost, c.host)
	}
}
