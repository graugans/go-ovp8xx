package ovp8xx

import (
	"encoding/xml"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
	const reply = `{"device":{"name":"test"}}`
	conf := *NewConfig(WitJSONString(`{"a":1}`))

	tests := []struct {
		name string
		do   func(c *Client) (string, error)
		want rpcCall
	}{
		{"Get", func(c *Client) (string, error) {
			cfg, err := c.Get([]string{"/device"})
			return cfg.String(), err
		}, rpcCall{efectorPath, "get", []string{"/device"}}},
		{"Set", func(c *Client) (string, error) {
			return "", c.Set(conf)
		}, rpcCall{efectorPath, "set", []string{conf.String()}}},
		{"GetInit", func(c *Client) (string, error) {
			cfg, err := c.GetInit()
			return cfg.String(), err
		}, rpcCall{efectorPath, "getInit", nil}},
		{"SaveInit all", func(c *Client) (string, error) {
			return "", c.SaveInit(nil)
		}, rpcCall{efectorPath, "saveInit", nil}},
		{"SaveInit pointers", func(c *Client) (string, error) {
			return "", c.SaveInit([]string{"/device", "/ports"})
		}, rpcCall{efectorPath, "saveInit", []string{"/device", "/ports"}}},
		{"FactoryReset", func(c *Client) (string, error) {
			return "", c.FactoryReset(true)
		}, rpcCall{efectorPath, "factoryReset", []string{"1"}}},
		{"GetSchema", func(c *Client) (string, error) {
			return c.GetSchema([]string{"/device"})
		}, rpcCall{efectorPath, "getSchema", []string{"/device"}}},
		{"Reboot", func(c *Client) (string, error) {
			return "", c.Reboot()
		}, rpcCall{efectorPath, "reboot", nil}},
		{"RebootToSWUpdate", func(c *Client) (string, error) {
			return "", c.RebootToSWUpdate()
		}, rpcCall{efectorPath, "rebootToRecovery", nil}},
		{"Remove", func(c *Client) (string, error) {
			return "", c.Remove("/applications/instances/app0")
		}, rpcCall{efectorPath, "remove", []string{"/applications/instances/app0"}}},
		{"GetFiltered", func(c *Client) (string, error) {
			cfg, err := c.GetDiagnosticClient().GetFiltered(conf)
			return cfg.String(), err
		}, rpcCall{diagnosticPath, "getFiltered", []string{conf.String()}}},
		{"GetFilterSchema", func(c *Client) (string, error) {
			cfg, err := c.GetDiagnosticClient().GetFilterSchema()
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
	_, err := NewClient(WithHost(host)).Get([]string{"/device"})
	assert.Error(t, err)
}
