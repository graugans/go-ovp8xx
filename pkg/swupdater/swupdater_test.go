package swupdater

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// newFakeSWUpdate starts a fake SWUpdate service. The upload endpoint answers
// with uploadStatus. After a successful upload the websocket reports success.
func newFakeSWUpdate(t *testing.T, uploadStatus int, listener net.Listener) (host string, port uint16) {
	t.Helper()
	uploaded := make(chan struct{}, 1)
	upgrader := websocket.Upgrader{}
	mux := http.NewServeMux()
	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(uploadStatus)
		if uploadStatus == http.StatusOK {
			uploaded <- struct{}{}
		}
	})
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer func() { _ = ws.Close() }()
		select {
		case <-uploaded:
			_ = ws.WriteJSON(map[string]string{"type": "message", "text": "SWUPDATE successful !"})
		case <-r.Context().Done():
		}
		// keep the connection open until the client closes it
		_, _, _ = ws.ReadMessage()
	})
	srv := httptest.NewUnstartedServer(mux)
	if listener != nil {
		srv.Listener = listener
	}
	srv.Start()
	t.Cleanup(srv.Close)
	h, p, err := net.SplitHostPort(srv.Listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	pn, _ := strconv.Atoi(p)
	return h, uint16(pn)
}

func firmwareFile(t *testing.T) string {
	t.Helper()
	name := filepath.Join(t.TempDir(), "firmware.swu")
	if err := os.WriteFile(name, []byte("firmware"), 0o600); err != nil {
		t.Fatal(err)
	}
	return name
}

func TestUpdateWithoutNotifications(t *testing.T) {
	host, port := newFakeSWUpdate(t, http.StatusOK, nil)
	swu := NewSWUpdater(host, port, nil)
	if err := swu.Update(firmwareFile(t), time.Second, 5*time.Second); err != nil {
		t.Fatalf("expected a successful update, got: %v", err)
	}
}

func TestUpdateWithNotifications(t *testing.T) {
	host, port := newFakeSWUpdate(t, http.StatusOK, nil)
	notifications := make(chan SWUpdaterNotification, 100)
	swu := NewSWUpdater(host, port, notifications)
	if err := swu.Update(firmwareFile(t), time.Second, 5*time.Second); err != nil {
		t.Fatalf("expected a successful update, got: %v", err)
	}
	count := 0
	for range notifications { // the channel is closed after the update
		count++
	}
	if count == 0 {
		t.Error("expected notifications")
	}
}

func TestUpdateUploadRejected(t *testing.T) {
	host, port := newFakeSWUpdate(t, http.StatusBadRequest, nil)
	swu := NewSWUpdater(host, port, nil)
	start := time.Now()
	err := swu.Update(firmwareFile(t), time.Second, 5*time.Second)
	if err == nil {
		t.Fatal("expected an error when the upload is rejected")
	}
	if time.Since(start) > 2*time.Second {
		t.Errorf("a rejected upload must fail immediately, not after the timeout: %v", err)
	}
}

func TestURLs(t *testing.T) {
	tests := []struct{ host, upload, status string }{
		{"192.168.0.69", "http://192.168.0.69:8080/upload", "ws://192.168.0.69:8080/ws"},
		{"::1", "http://[::1]:8080/upload", "ws://[::1]:8080/ws"},
		{"fe80::1%eth0", "http://[fe80::1%25eth0]:8080/upload", "ws://[fe80::1%25eth0]:8080/ws"},
	}
	for _, tt := range tests {
		s := NewSWUpdater(tt.host, 8080, nil)
		if s.urlUpload != tt.upload || s.urlStatus != tt.status {
			t.Errorf("host %q: got %q and %q, want %q and %q", tt.host, s.urlUpload, s.urlStatus, tt.upload, tt.status)
		}
	}
}

func TestUpdateIPv6(t *testing.T) {
	l, err := net.Listen("tcp", "[::1]:0")
	if err != nil {
		t.Skip("IPv6 loopback not available:", err)
	}
	host, port := newFakeSWUpdate(t, http.StatusOK, l)
	if host != "::1" {
		t.Fatalf("unexpected host %q", host)
	}
	swu := NewSWUpdater(host, port, nil)
	if err := swu.Update(firmwareFile(t), time.Second, 5*time.Second); err != nil {
		t.Fatalf("expected a successful update via IPv6, got: %v", err)
	}
}

// newFakeDeviceInfo starts a fake SWUpdate service which answers the device
// information requests with the status returned by status. The argument of
// status is the number of the request, starting at 1.
func newFakeDeviceInfo(t *testing.T, status func(request int32) int) (*SWUpdater, *atomic.Int32) {
	t.Helper()
	requests := &atomic.Int32{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/device-info" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(status(requests.Add(1)))
	}))
	t.Cleanup(srv.Close)
	h, p, err := net.SplitHostPort(srv.Listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	pn, _ := strconv.Atoi(p)
	return NewSWUpdater(h, uint16(pn), nil), requests
}

func fastAvailabilityPolling(t *testing.T) {
	t.Helper()
	old := availabilityPollInterval
	availabilityPollInterval = 10 * time.Millisecond
	t.Cleanup(func() { availabilityPollInterval = old })
}

func timeoutCtx(t *testing.T, d time.Duration) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), d)
	t.Cleanup(cancel)
	return ctx
}

func TestIsAvailable(t *testing.T) {
	fastAvailabilityPolling(t)
	swu, requests := newFakeDeviceInfo(t, func(int32) int { return http.StatusOK })
	ok, err := swu.IsAvailable(timeoutCtx(t, 2*time.Second))
	if !ok || err != nil {
		t.Fatalf("expected the service to be available, got %v, %v", ok, err)
	}
	if n := requests.Load(); n != requiredAvailabilityChecks {
		t.Errorf("expected %d checks, got %d", requiredAvailabilityChecks, n)
	}
}

func TestIsAvailableRequiresConsecutiveChecks(t *testing.T) {
	fastAvailabilityPolling(t)
	// Two good answers followed by a failure must not count
	swu, requests := newFakeDeviceInfo(t, func(request int32) int {
		if request == 3 {
			return http.StatusServiceUnavailable
		}
		return http.StatusOK
	})
	ok, err := swu.IsAvailable(timeoutCtx(t, 2*time.Second))
	if !ok || err != nil {
		t.Fatalf("expected the service to be available, got %v, %v", ok, err)
	}
	if n := requests.Load(); n != 6 {
		t.Errorf("expected 6 checks, got %d", n)
	}
}

func TestIsAvailableTimeout(t *testing.T) {
	fastAvailabilityPolling(t)
	swu, _ := newFakeDeviceInfo(t, func(int32) int { return http.StatusNotFound })
	ok, err := swu.IsAvailable(timeoutCtx(t, 100*time.Millisecond))
	if ok || err == nil {
		t.Fatalf("expected a timeout, got %v, %v", ok, err)
	}
}
