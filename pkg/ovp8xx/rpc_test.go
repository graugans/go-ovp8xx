package ovp8xx

import (
	"context"
	"encoding/xml"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// rpcCall is a single XML-RPC request received by the fake device.
type rpcCall struct {
	path   string
	method string
	params []string // the string and boolean parameter values in order
}

// newRecordingDevice starts an XML-RPC server which records every request and
// answers it with reply as string, or with an XML-RPC fault if fault is set.
func newRecordingDevice(t *testing.T, reply string, fault bool) (*Client, *[]rpcCall) {
	t.Helper()
	var calls []rpcCall
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		var req struct {
			Method string   `xml:"methodName"`
			Values []string `xml:"params>param>value>string"`
			Arrays []string `xml:"params>param>value>array>data>value>string"`
			Bools  []string `xml:"params>param>value>boolean"`
		}
		require.NoError(t, xml.Unmarshal(body, &req))
		params := append(append(req.Values, req.Arrays...), req.Bools...)
		calls = append(calls, rpcCall{path: r.URL.Path, method: req.Method, params: params})

		w.Header().Set("Content-Type", "text/xml")
		if fault {
			_, _ = w.Write([]byte(`<?xml version="1.0"?><methodResponse><fault><value><struct>` +
				`<member><name>faultCode</name><value><int>101000</int></value></member>` +
				`<member><name>faultString</name><value><string>boom</string></value></member>` +
				`</struct></value></fault></methodResponse>`))
			return
		}
		_, _ = w.Write([]byte(`<?xml version="1.0"?><methodResponse><params><param><value><string>` +
			reply + `</string></value></param></params></methodResponse>`))
	}))
	t.Cleanup(srv.Close)
	return NewClient(WithHost(strings.TrimPrefix(srv.URL, "http://"))), &calls
}

const (
	efectorPath    = "/api/rpc/v1/com.ifm.efector/"
	diagnosticPath = "/api/rpc/v1/com.ifm.diagnostic/"
)

func TestRPCMethods(t *testing.T) {
	ctx := context.Background()
	const reply = `{"device":{"name":"test"}}`
	conf := *NewConfig(WithJSONString(`{"a":1}`))

	tests := []struct {
		name string
		do   func(c *Client) (string, error)
		want rpcCall
	}{
		{"Get", func(c *Client) (string, error) {
			cfg, err := c.Get(ctx, []string{"/device"})
			return cfg.String(), err
		}, rpcCall{efectorPath, "get", []string{"/device"}}},
		{"Set", func(c *Client) (string, error) {
			return "", c.Set(ctx, conf)
		}, rpcCall{efectorPath, "set", []string{conf.String()}}},
		{"GetInit", func(c *Client) (string, error) {
			cfg, err := c.GetInit(ctx)
			return cfg.String(), err
		}, rpcCall{efectorPath, "getInit", nil}},
		{"SaveInit all", func(c *Client) (string, error) {
			return "", c.SaveInit(ctx, nil)
		}, rpcCall{efectorPath, "saveInit", nil}},
		{"SaveInit pointers", func(c *Client) (string, error) {
			return "", c.SaveInit(ctx, []string{"/device", "/ports"})
		}, rpcCall{efectorPath, "saveInit", []string{"/device", "/ports"}}},
		{"FactoryReset", func(c *Client) (string, error) {
			return "", c.FactoryReset(ctx, true)
		}, rpcCall{efectorPath, "factoryReset", []string{"1"}}},
		{"GetSchema", func(c *Client) (string, error) {
			return c.GetSchema(ctx, []string{"/device"})
		}, rpcCall{efectorPath, "getSchema", []string{"/device"}}},
		{"Reboot", func(c *Client) (string, error) {
			return "", c.Reboot(ctx)
		}, rpcCall{efectorPath, "reboot", nil}},
		{"RebootToSWUpdate", func(c *Client) (string, error) {
			return "", c.RebootToSWUpdate(ctx)
		}, rpcCall{efectorPath, "rebootToRecovery", nil}},
		{"Remove", func(c *Client) (string, error) {
			return "", c.Remove(ctx, "/applications/instances/app0")
		}, rpcCall{efectorPath, "remove", []string{"/applications/instances/app0"}}},
		{"GetFiltered", func(c *Client) (string, error) {
			cfg, err := c.GetDiagnosticClient().GetFiltered(ctx, conf)
			return cfg.String(), err
		}, rpcCall{diagnosticPath, "getFiltered", []string{conf.String()}}},
		{"GetFilterSchema", func(c *Client) (string, error) {
			cfg, err := c.GetDiagnosticClient().GetFilterSchema(ctx)
			return cfg.String(), err
		}, rpcCall{diagnosticPath, "getFilterSchema", nil}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, calls := newRecordingDevice(t, reply, false)
			got, err := tt.do(client)
			require.NoError(t, err)
			if got != "" {
				assert.JSONEq(t, reply, got)
			}
			require.Len(t, *calls, 1)
			assert.Equal(t, tt.want.path, (*calls)[0].path)
			assert.Equal(t, tt.want.method, (*calls)[0].method)
			assert.ElementsMatch(t, tt.want.params, (*calls)[0].params)
		})

		t.Run(tt.name+" fault", func(t *testing.T) {
			client, _ := newRecordingDevice(t, "", true)
			_, err := tt.do(client)
			assert.ErrorContains(t, err, "boom")
		})
	}
}

func TestRPCConnectionRefused(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	host := strings.TrimPrefix(srv.URL, "http://")
	srv.Close()
	_, err := NewClient(WithHost(host)).Get(context.Background(), []string{"/device"})
	assert.Error(t, err)
}

// newHangingDevice starts a server which never answers until the test ends.
func newHangingDevice(t *testing.T) string {
	t.Helper()
	done := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-done:
		case <-r.Context().Done():
		}
	}))
	t.Cleanup(srv.Close)
	t.Cleanup(func() { close(done) })
	return strings.TrimPrefix(srv.URL, "http://")
}

func TestRPCContextCanceled(t *testing.T) {
	client := NewClient(WithHost(newHangingDevice(t)))
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, err := client.Get(ctx, []string{"/device"})
	require.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Less(t, time.Since(start), 2*time.Second)
}

func TestRPCTimeout(t *testing.T) {
	client := NewClient(WithHost(newHangingDevice(t)), WithTimeout(100*time.Millisecond))
	start := time.Now()
	// The diagnostic client inherits the timeout
	_, err := client.GetDiagnosticClient().GetFilterSchema(context.Background())
	require.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Less(t, time.Since(start), 2*time.Second)
}

func TestWithHTTPClient(t *testing.T) {
	var used bool
	hc := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		used = true
		return http.DefaultTransport.RoundTrip(r)
	})}
	client, _ := newRecordingDevice(t, `{}`, false)
	client = NewClient(WithHost(client.host), WithHTTPClient(hc))
	_, err := client.Get(context.Background(), []string{"/device"})
	require.NoError(t, err)
	assert.True(t, used, "custom HTTP client not used")
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
