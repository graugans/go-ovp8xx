package pcic_test

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/graugans/go-ovp8xx/v2/pkg/pcic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeDevice reads PCIC requests and answers each of them with the same
// ticket, echoing the content. An optional delay is applied before answering.
type fakeDevice struct {
	requests  *bufio.Reader
	responses io.Writer
	delay     func() time.Duration
	mu        sync.Mutex
}

func (d *fakeDevice) serve(t *testing.T) {
	for {
		header := make([]byte, 20)
		if _, err := io.ReadFull(d.requests, header); err != nil {
			return
		}
		ticket := string(header[:4])
		length, err := strconv.Atoi(string(header[5:14]))
		if err != nil || string(header[16:20]) != ticket {
			t.Errorf("malformed request header: %q", header)
			return
		}
		body := make([]byte, length-4)
		if _, err := io.ReadFull(d.requests, body); err != nil {
			return
		}
		content := body[:len(body)-2]
		go func() {
			if d.delay != nil {
				time.Sleep(d.delay())
			}
			d.mu.Lock()
			defer d.mu.Unlock()
			_, _ = d.responses.Write(pcicMessage(ticket, content))
		}()
	}
}

// newClientWithDevice connects a PCICClient to a fakeDevice and starts the
// receive loop.
func newClientWithDevice(t *testing.T, delay func() time.Duration) *pcic.PCICClient {
	t.Helper()
	reqR, reqW := io.Pipe()
	respR, respW := io.Pipe()
	t.Cleanup(func() {
		_ = reqW.Close()
		_ = respW.Close()
	})
	dev := &fakeDevice{requests: bufio.NewReader(reqR), responses: respW, delay: delay}
	go dev.serve(t)

	p, err := pcic.NewPCICClient(pcic.WithBufioReaderWriter(
		bufio.NewReadWriter(bufio.NewReader(respR), bufio.NewWriter(reqW)),
	))
	require.NoError(t, err)
	go func() {
		for p.ProcessIncomming(&recordingHandler{}) == nil {
		}
	}()
	return p
}

func TestConcurrentSend(t *testing.T) {
	p := newClientWithDevice(t, nil)
	var wg sync.WaitGroup
	for g := 0; g < 16; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				cmd := []byte(fmt.Sprintf("cmd-%d-%d-%s", g, i, bytes.Repeat([]byte("x"), 1000)))
				res, err := p.Send(ctx, cmd)
				cancel()
				if !assert.NoError(t, err) || !assert.Equal(t, cmd, res) {
					return
				}
			}
		}(g)
	}
	wg.Wait()
}

func TestSendTimeoutDoesNotBlockReceiver(t *testing.T) {
	// Responses arrive around the time the requests time out. A late response
	// must neither block the receive loop nor subsequent requests.
	p := newClientWithDevice(t, func() time.Duration {
		return time.Duration(rand.Intn(300)) * time.Microsecond
	})
	done := make(chan error, 1)
	go func() {
		for i := 0; i < 500; i++ {
			ctx, cancel := context.WithTimeout(context.Background(), 150*time.Microsecond)
			_, _ = p.Send(ctx, []byte("late"))
			cancel()
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_, err := p.Send(ctx, []byte("final"))
		done <- err
	}()
	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("Send is blocked, the receive loop dead locked")
	}
}
