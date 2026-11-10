package swupdater

import (
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// newFakeSWUpdate starts a fake SWUpdate service. The upload endpoint answers
// with uploadStatus. After a successful upload the websocket reports success.
func newFakeSWUpdate(t *testing.T, uploadStatus int) (host string, port uint16) {
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
	srv := httptest.NewServer(mux)
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
	host, port := newFakeSWUpdate(t, http.StatusOK)
	swu := NewSWUpdater(host, port, nil)
	if err := swu.Update(firmwareFile(t), time.Second, 5*time.Second); err != nil {
		t.Fatalf("expected a successful update, got: %v", err)
	}
}

func TestUpdateWithNotifications(t *testing.T) {
	host, port := newFakeSWUpdate(t, http.StatusOK)
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
	host, port := newFakeSWUpdate(t, http.StatusBadRequest)
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
