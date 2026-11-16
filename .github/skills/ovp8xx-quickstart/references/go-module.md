# go-ovp8xx Module Reference

Module path: `github.com/graugans/go-ovp8xx/v2` (MIT licensed, personal
project, not an official ifm product).

## Packages

| Package | Purpose |
|---------|---------|
| `pkg/ovp8xx` | XML-RPC client for device configuration (get/set/schema/diagnostics) |
| `pkg/pcic` | PCIC protocol: TCP result stream with frames and chunks (3D/2D/IMU data) |
| `pkg/swupdater` | Firmware update via websocket |

## Configuration client (`pkg/ovp8xx`)

```go
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/graugans/go-ovp8xx/v2/pkg/ovp8xx"
)

func main() {
	host := os.Getenv("OVP8XX_IP")
	if host == "" {
		host = ovp8xx.DefaultHost // "192.168.0.69"
	}
	o3r := ovp8xx.NewClient(
		ovp8xx.WithHost(host),
		ovp8xx.WithTimeout(10*time.Second),
	)
	ctx := context.Background()

	// Scoped read with JSON pointers; empty slice reads the full document.
	conf, err := o3r.Get(ctx, []string{"/device/swVersion"})
	if err != nil {
		panic(err)
	}
	fmt.Println(conf.String())
}
```

Key `Client` methods (all take a `context.Context`):

- `Get(ctx, pointers []string) (Config, error)` — scoped configuration read
- `Set(ctx, conf Config) error` — write; build the payload with
  `ovp8xx.NewConfig(ovp8xx.WithJSONString("{\"device\":{...}}"))`
- `GetSchema(ctx, pointers []string) (string, error)` — live JSON Schema
- `GetInit` / `SaveInit` — boot-persistent configuration
- `GetDiagnosticClient().GetFiltered(ctx, conf Config) (Config, error)` — diagnostics
- `IsAvailable(ctx, opts...)` — wait helpers `AndPortsAreOnline()`, `AndAppsAreOnline()`
- `Reboot`, `RebootToSWUpdate`, `FactoryReset` — state-changing, use deliberately

`Config` wraps a JSON document (`Config{JSON string}`, `String()` returns it).
Use `ovp8xx.IsTimeoutError(err)` to detect timeouts.

## PCIC result streaming (`pkg/pcic`)

Each sensor port exposes its own PCIC TCP port (read
`/ports/portX/data/pcicTCPPort` from the configuration; `pcic.DefaultPort` is
50010). Implement the `MessageHandler` interface and pump the stream:

```go
type handler struct{}

func (h *handler) Result(f pcic.Frame) {
	fmt.Printf("frame with %d chunks\n", len(f.Chunks))
}
func (h *handler) Error(e pcic.ErrorMessage)               {}
func (h *handler) Notification(n pcic.NotificationMessage) {}

func stream(host string, port uint16) error {
	client, err := pcic.NewPCICClient(pcic.WithTCPClient(host, port))
	if err != nil {
		return err
	}
	defer client.Close()
	for {
		if err := client.ProcessIncomming(&handler{}); err != nil {
			return err
		}
	}
}
```

Chunk types include `RADIAL_DISTANCE_COMPRESSED`, `AMPLITUDE_COMPRESSED`,
`CONFIDENCE`, `TOF_INFO`, `O3R_RESULT_IMU` (see `pcic.ChunkType` constants).

## Conventions

- Device address: environment variable `OVP8XX_IP`, overridable by a `--ip` flag
- Set `NO_PROXY`/`no_proxy` for the device IP when HTTP proxies are configured
- The port must be in state `RUN` to produce PCIC frames; check
  `/ports/portX/state` via `Get`
