package mcpserver

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func connect(t *testing.T, server *mcp.Server) (*mcp.ClientSession, context.Context) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	st, ct := mcp.NewInMemoryTransports()
	ss, err := server.Connect(ctx, st, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = ss.Close() })
	cs, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil).Connect(ctx, ct, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = cs.Close() })
	return cs, ctx
}

func call(t *testing.T, ctx context.Context, cs *mcp.ClientSession, name string, args any) *mcp.CallToolResult {
	t.Helper()
	result, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
	require.NoError(t, err)
	return result
}

func TestServer(t *testing.T) {
	requests := make(chan string, 20)
	device := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
			return
		}
		requests <- string(body)
		var req struct {
			Method string `xml:"methodName"`
		}
		if err := xml.Unmarshal(body, &req); err != nil {
			t.Error(err)
			return
		}
		reply := `{"device":{"status":"OPERATE"},"ports":{"portX":{"info":{"features":{"type":"3D"}},"state":"RUN","mode":"test","data":{"pcicTCPPort":50123,"availablePCICOutput":["CONFIDENCE"]},"acquisition":{"framerate":10}}},"applications":{"classes":{"test":{"license":"FREE","version":{"major":1},"instanceSchema":{"large":true}}}}}`
		switch req.Method {
		case "get", "getSchema":
		case "getFiltered":
			require.Equal(t, "/api/rpc/v1/com.ifm.diagnostic/", r.URL.Path)
			reply = `{"events":[],"groups":{"portX":"no_incident"},"timestamp":"1794579722996595250"}`
		case "getFilterSchema":
			reply = `{"type":"object"}`
		default:
			t.Errorf("unexpected device mutation/method: %s", req.Method)
		}
		w.Header().Set("Content-Type", "text/xml")
		_, _ = io.WriteString(w, `<methodResponse><params><param><value><string>`)
		_ = xml.EscapeText(w, []byte(reply))
		_, _ = io.WriteString(w, `</string></value></param></params></methodResponse>`)
	}))
	defer device.Close()
	cs, ctx := connect(t, New(strings.TrimPrefix(device.URL, "http://"), time.Second))
	tools, err := cs.ListTools(ctx, nil)
	require.NoError(t, err)
	require.Len(t, tools.Tools, 4)
	for _, tool := range tools.Tools {
		require.NotNil(t, tool.Annotations)
		require.True(t, tool.Annotations.ReadOnlyHint)
	}

	for _, name := range []string{"get_config", "get_schema"} {
		result := call(t, ctx, cs, name, map[string]any{"paths": []string{"/device/status"}})
		require.False(t, result.IsError, "%v", result.Content)
		require.Contains(t, <-requests, "/device/status")
		require.Contains(t, result.Content[0].(*mcp.TextContent).Text, `"device"`)
		result = call(t, ctx, cs, name, map[string]any{})
		require.False(t, result.IsError)
		<-requests
	}
	for _, tc := range []struct {
		args any
		want string
	}{
		{map[string]any{}, `{"state":"active"}`},
		{map[string]any{"filter": map[string]any{}}, `{}`},
		{map[string]any{"filter": []any{map[string]any{"severity": "major"}}}, `[{"severity":"major"}]`},
	} {
		result := call(t, ctx, cs, "get_diagnostics", tc.args)
		require.False(t, result.IsError)
		var req struct {
			Value string `xml:"params>param>value>string"`
		}
		require.NoError(t, xml.Unmarshal([]byte(<-requests), &req))
		require.JSONEq(t, tc.want, req.Value)
	}

	result := call(t, ctx, cs, "get_device_overview", map[string]any{})
	require.False(t, result.IsError)
	<-requests
	require.Contains(t, <-requests, "getFiltered")
	text := result.Content[0].(*mcp.TextContent).Text
	require.Contains(t, text, `"portX"`)
	require.Contains(t, text, `50123`)
	require.NotContains(t, text, "instanceSchema")
	require.NotContains(t, text, "acquisition")
	require.Contains(t, text, "observedAt")
	require.Contains(t, text, "no_incident")

	resources, err := cs.ListResources(ctx, nil)
	require.NoError(t, err)
	require.Len(t, resources.Resources, 3)
	for _, res := range resources.Resources {
		result, err := cs.ReadResource(ctx, &mcp.ReadResourceParams{URI: res.URI})
		require.NoError(t, err)
		require.Len(t, result.Contents, 1)
		require.NotEmpty(t, result.Contents[0].Text)
	}
	require.Contains(t, <-requests, "getFilterSchema")

	for _, args := range []any{
		map[string]any{"paths": []string{"device/status"}},
		map[string]any{"paths": []string{"/bad~2escape"}},
	} {
		require.True(t, call(t, ctx, cs, "get_config", args).IsError)
	}
	require.True(t, call(t, ctx, cs, "get_diagnostics", map[string]any{"filter": "bad"}).IsError)
	_, err = cs.CallTool(ctx, &mcp.CallToolParams{Name: "set", Arguments: map[string]any{}})
	require.Error(t, err)
	require.Empty(t, requests)
}

func TestDeviceFailures(t *testing.T) {
	for _, mode := range []string{"invalid JSON", "timeout", "fault"} {
		t.Run(mode, func(t *testing.T) {
			device := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/xml")
				switch mode {
				case "timeout":
					select {
					case <-r.Context().Done():
					case <-time.After(200 * time.Millisecond):
					}
				case "fault":
					_, _ = io.WriteString(w, `<methodResponse><fault><value><struct><member><name>faultCode</name><value><int>1</int></value></member><member><name>faultString</name><value><string>device fault</string></value></member></struct></value></fault></methodResponse>`)
				default:
					_, _ = io.WriteString(w, `<methodResponse><params><param><value><string>not JSON</string></value></param></params></methodResponse>`)
				}
			}))
			defer device.Close()
			cs, ctx := connect(t, New(strings.TrimPrefix(device.URL, "http://"), 50*time.Millisecond))
			for _, name := range []string{"get_config", "get_schema", "get_diagnostics", "get_device_overview"} {
				result := call(t, ctx, cs, name, map[string]any{})
				require.True(t, result.IsError, "expected %s error for %s", mode, name)
				require.NotEmpty(t, result.Content)
			}
			_, err := cs.ReadResource(ctx, &mcp.ReadResourceParams{URI: "device://diagnostics/filter-schema"})
			require.Error(t, err)
		})
	}
}

func TestJSONPreservesNumbers(t *testing.T) {
	value, err := decodeJSON(`{"timestamp":1794579722996595250}`, nil)
	require.NoError(t, err)
	encoded, err := json.Marshal(value)
	require.NoError(t, err)
	require.Contains(t, string(encoded), "1794579722996595250")
}

// This opt-in test verifies the real CLI transport without changing the device.
func TestLiveStdio(t *testing.T) {
	host := os.Getenv("OVP8XX_MCP_TEST_HOST")
	if host == "" {
		t.Skip("set OVP8XX_MCP_TEST_HOST to run read-only hardware checks")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "run", "./cmd/ovp8xx", "mcp", "--ip", host)
	cmd.Dir = "../.."
	cmd.Stderr = os.Stderr
	cs, err := mcp.NewClient(&mcp.Implementation{Name: "live-test", Version: "1"}, nil).Connect(ctx, &mcp.CommandTransport{Command: cmd}, nil)
	require.NoError(t, err)
	defer func() { require.NoError(t, cs.Close()) }()
	tools, err := cs.ListTools(ctx, nil)
	require.NoError(t, err)
	require.Len(t, tools.Tools, 4)
	for _, tc := range []struct {
		name string
		args any
	}{
		{"get_device_overview", map[string]any{}},
		{"get_config", map[string]any{"paths": []string{"/device/swVersion", "/device/diagnostic/confInitStages"}}},
		{"get_schema", map[string]any{"paths": []string{"/ports/port0/state"}}},
		{"get_diagnostics", map[string]any{}},
	} {
		res := call(t, ctx, cs, tc.name, tc.args)
		require.False(t, res.IsError, "%v", res.Content)
		t.Logf("%s: %s", tc.name, res.Content[0].(*mcp.TextContent).Text)
	}
	resources, err := cs.ListResources(ctx, nil)
	require.NoError(t, err)
	for _, res := range resources.Resources {
		_, err := cs.ReadResource(ctx, &mcp.ReadResourceParams{URI: res.URI})
		require.NoError(t, err, fmt.Sprintf("read %s", res.URI))
	}
}

// xmlRPCCall is a decoded XML-RPC request of the fake device.
type xmlRPCCall struct {
	Method string `xml:"methodName"`
	Body   string
}

// newWriteDevice starts a fake device which records all calls and answers
// get requests with reply.
func newWriteDevice(t *testing.T, reply string) (string, chan xmlRPCCall) {
	t.Helper()
	calls := make(chan xmlRPCCall, 20)
	device := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
			return
		}
		var c xmlRPCCall
		if err := xml.Unmarshal(body, &c); err != nil {
			t.Error(err)
			return
		}
		c.Body = string(body)
		calls <- c
		w.Header().Set("Content-Type", "text/xml")
		if c.Method != "get" {
			_, _ = io.WriteString(w, `<methodResponse><params><param><value><string></string></value></param></params></methodResponse>`)
			return
		}
		_, _ = io.WriteString(w, `<methodResponse><params><param><value><string>`)
		_ = xml.EscapeText(w, []byte(reply))
		_, _ = io.WriteString(w, `</string></value></param></params></methodResponse>`)
	}))
	t.Cleanup(device.Close)
	return strings.TrimPrefix(device.URL, "http://"), calls
}

func TestWriteTools(t *testing.T) {
	host, calls := newWriteDevice(t, `{"device":{"info":{"name":"front"}}}`)
	cs, ctx := connect(t, New(host, time.Second, WithWrite()))

	tools, err := cs.ListTools(ctx, nil)
	require.NoError(t, err)
	require.Len(t, tools.Tools, 6)
	for _, tool := range tools.Tools {
		switch tool.Name {
		case "set_config", "save_init":
			require.False(t, tool.Annotations.ReadOnlyHint)
			require.NotNil(t, tool.Annotations.DestructiveHint)
			require.True(t, *tool.Annotations.DestructiveHint)
		default:
			require.True(t, tool.Annotations.ReadOnlyHint, tool.Name)
		}
	}
	api, err := cs.ReadResource(ctx, &mcp.ReadResourceParams{URI: "device://knowledge/api"})
	require.NoError(t, err)
	require.Contains(t, api.Contents[0].Text, "save_init")
	require.NotContains(t, api.Contents[0].Text, "read-only")

	// The fragment is forwarded unchanged, large integers must not be rounded
	result := call(t, ctx, cs, "set_config", map[string]any{"config": json.RawMessage(
		`{"device":{"info":{"name":"front","n":9007199254740993}},"ports":{"port0":{"a/b":{}}}}`)})
	require.False(t, result.IsError, "%v", result.Content)
	set := <-calls
	require.Equal(t, "set", set.Method)
	require.Contains(t, set.Body, "9007199254740993")
	get := <-calls
	require.Equal(t, "get", get.Method)
	for _, path := range []string{"/device/info/n", "/device/info/name", "/ports/port0/a~1b"} {
		require.Contains(t, get.Body, path)
	}
	text := result.Content[0].(*mcp.TextContent).Text
	require.Contains(t, text, `"readBack"`)
	require.Contains(t, text, `"persisted":false`)

	result = call(t, ctx, cs, "save_init", map[string]any{"paths": []string{"/device/info"}})
	require.False(t, result.IsError, "%v", result.Content)
	save := <-calls
	require.Equal(t, "saveInit", save.Method)
	require.Contains(t, save.Body, "/device/info")

	// Invalid input must not reach the device
	for _, tc := range []struct {
		name string
		args any
	}{
		{"set_config", map[string]any{"config": map[string]any{}}},
		{"save_init", map[string]any{"paths": []string{}}},
		{"save_init", map[string]any{"paths": []string{"device"}}},
	} {
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: tc.name, Arguments: tc.args})
		require.True(t, err != nil || res.IsError, "%s %v", tc.name, tc.args)
	}
	require.Empty(t, calls)
}

func TestLeafPointers(t *testing.T) {
	paths, err := leafPointers(json.RawMessage(`{"b":{"c":1,"d":[1,2]},"a~":"x","e":{}}`))
	require.NoError(t, err)
	require.Equal(t, []string{"/a~0", "/b/c", "/b/d", "/e"}, paths)
	for _, bad := range []string{`[]`, `{}`, `1`, `{`} {
		_, err := leafPointers(json.RawMessage(bad))
		require.Error(t, err, bad)
	}
}
