

# Go client for the ifm OVP8xx series of devices

⚠️ This is my personal project it is not an official ifm product. Please use the official [ifm3d](https://github.com/ifm/ifm3d) project if you need an officially supported version ⚠️

## Introduction

A GO module and cli to access the ifm OVP8xx series of devices.


[![GitHub go.mod Go version of a Go module](https://img.shields.io/github/go-mod/go-version/graugans/go-ovp8xx.svg)](https://github.com/graugans/go-ovp8xx)
[![Go Reference](https://pkg.go.dev/badge/github.com/graugans/go-ovp8xx/v2.svg)](https://pkg.go.dev/github.com/graugans/go-ovp8xx/v2)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
![example workflow](https://github.com/graugans/go-ovp8xx/actions/workflows/go.yml/badge.svg)
[![codecov](https://codecov.io/gh/graugans/go-ovp8xx/graph/badge.svg?token=BU6UPYCUPI)](https://codecov.io/gh/graugans/go-ovp8xx)

## Project status

This project is still a work in progress and will suffer from breaking API changes. Please be warned. In case you have any suggestions or want to contribute please feel free to open an issue or pull request.

## CLI

One of the benefits of the Go language is the easy way of producing statically linked binaries to be used on all major platforms. One of the `ovp8xx` core features is the CLI interface. The design tries to stay as close as possible to the XML-RPC API.

With the CLI you can get the configuration from the device and [filter](doc/filter.md) it as you need. After transforming the config it can be written back to the device.

### Supported device families (multicall binary)

Besides the OVP8xx the API compatible OVP9xx and O3Cxxx devices are supported as well. Similar to [BusyBox](https://busybox.net/) the `ovp8xx` binary is a multicall binary: the device family is selected by the name the binary is invoked with. The default is `ovp8xx`, to use the other variants create a link:

```sh
ln -s ovp8xx ovp9xx
ln -s ovp8xx o3cxxx
```

| Binary name | Device family | IP environment variable |
|-------------|---------------|-------------------------|
| `ovp8xx`    | OVP8xx        | `OVP8XX_IP`             |
| `ovp9xx`    | OVP9xx        | `OVP9XX_IP`, falls back to `OVP8XX_IP` |
| `o3cxxx`    | O3Cxxx        | `O3CXXX_IP`, falls back to `OVP8XX_IP` |

On Windows the binary can simply be copied or renamed (e.g. `o3cxxx.exe`). If renaming or linking is not an option, the variant can also be selected with the environment variable `OVP8XX_VARIANT` (e.g. `OVP8XX_VARIANT=ovp9xx`). A link name (`ovp9xx`, `o3cxxx`) takes precedence over the environment variable.

### CLI PCIC Receiver

There is a simple PCIC data receiver which is useful to check the data transferred over PCIC. PCIC is a proprietary protocol designed by ifm. For the OVP8xx line of devices the results provided by PCIC are encapsulated into a Chunk. The following information is available for each Chunk:

- **FrameCount**, This counts the frames received within this connection.
- **Index**, The index of the Chunk within the frame.
- **Type**, The unique Chunk type
- **Size**, The size in bytes of the Chunk
- **Status**, The the device default: 0
- **Timestamp**, The timestamp of the frame
- **Bytes**, The payload

Typically only the frame count is printed out. To dump more information please use `--dump` to receive a full representation of the Chunk. This can be filtered by the Go `text/template` notation like this:

```sh
ovp8xx pcic --dump  --template "Chunk {{.Index}}:\n  Type: {{.Type}}\n  Size: {{.Size}} bytes\n Data Hexdump:\n{{hexdump_range .Bytes 0 32}}"
```

There are two custom commands for the `text/template`:
- **hexdump <data>**, dumps the argument given as an Hexdump
- **hexdump_range <data> <offset> <length>**, dumps the data starting at an offset and a given length. If an length of `-1` is provided the content of `<data>` will be outputted starting at the `<offset>`

For linebreaks within the template string please use `\n`.

### CLI Installation

#### Install script (Linux and macOS)

The quickest way is the [install script](install.sh). It downloads the latest release for your OS and CPU architecture, verifies it against the `checksums.txt` of the release and installs `ovp8xx` together with the `ovp9xx` and `o3cxxx` links to `~/.local/bin`:

```sh
curl -fsSL https://raw.githubusercontent.com/graugans/go-ovp8xx/v2/install.sh | sh
```

The script never calls `sudo` on its own. Options are passed after `sh -s --`:

```sh
# install a specific release system wide
curl -fsSL https://raw.githubusercontent.com/graugans/go-ovp8xx/v2/install.sh | sudo sh -s -- --version v2.11.0 --bin-dir /usr/local/bin
```

| Option            | Environment variable | Default            | Description                          |
|-------------------|----------------------|--------------------|--------------------------------------|
| `-b, --bin-dir`   | `OVP8XX_INSTALL_DIR` | `$HOME/.local/bin` | Installation directory               |
| `-v, --version`   | `OVP8XX_VERSION`     | `latest`           | Release to install, e.g. `v2.11.0`   |
| `--no-links`      |                      |                    | Do not create the `ovp9xx` and `o3cxxx` links |

If you prefer to review the script before running it:

```sh
curl -fsSLO https://raw.githubusercontent.com/graugans/go-ovp8xx/v2/install.sh
less install.sh
sh install.sh
```

Running the same command again updates an existing installation.

#### Pre Build Binaries

Alternatively download the pre-build binary from the [GitHub Release page](https://github.com/graugans/go-ovp8xx/releases). This is also the way to go on Windows.

⚠️ The Windows binary maybe flagged by a Virus Scanner, please also read the note from the [Go Team](https://go.dev/doc/faq#virus) ⚠️

On Linux and macOS extract the archive and copy the binary into a directory within your `PATH`:

```sh
tar xzf go-ovp8xx_Linux_x86_64.tar.gz ovp8xx
sudo install -m 0755 ovp8xx /usr/local/bin/ovp8xx
```

Each release archive comes with a Software Bill of Materials (SBOM) in the [SPDX](https://spdx.dev/) JSON format, e.g. `go-ovp8xx_Linux_x86_64.tar.gz.sbom.json`. It lists all Go modules compiled into the binary and is covered by `checksums.txt`.

#### Go install

If you have a decent Go version installed

```sh
go install github.com/graugans/go-ovp8xx/v2/cmd/ovp8xx@latest
```

The binary is installed to `$(go env GOPATH)/bin`, make sure this directory is part of your `PATH`.

#### Enable the OVP9xx and O3Cxxx variants

The [install script](#install-script-linux-and-macos) creates these links automatically. When installing from the release archives or with `go install`, only the `ovp8xx` binary is available. To use the [other device families](#supported-device-families-multicall-binary) create links next to it.

Linux and macOS:

```sh
cd "$(dirname "$(command -v ovp8xx)")"
sudo ln -s ovp8xx ovp9xx
sudo ln -s ovp8xx o3cxxx
```

Drop the `sudo` if the binary lives in a user writable directory such as `$(go env GOPATH)/bin`.

Windows (PowerShell): symbolic links require administrator rights or the developer mode, so a hard link or a plain copy is the simpler choice:

```powershell
New-Item -ItemType HardLink -Path ovp9xx.exe -Target ovp8xx.exe
New-Item -ItemType HardLink -Path o3cxxx.exe -Target ovp8xx.exe
```

Verify the installation:

```sh
ovp9xx --help   # A command line application to interact with the ifm OVP9xx series of devices
```

#### Shell completion

The CLI generates completion scripts for bash, zsh, fish and PowerShell, which complete the commands and their flags.

To enable the completion for the current bash session:

```sh
. <(ovp8xx completion bash)
```

Add this line to your `~/.bashrc` to enable it permanently. The bash completion depends on the [bash-completion](https://github.com/scop/bash-completion) package, which is installed by default on most Linux distributions.

The other shells work the same way:

```sh
source <(ovp8xx completion zsh)                              # zsh, requires compinit to be loaded
ovp8xx completion fish | source                              # fish
ovp8xx completion powershell | Out-String | Invoke-Expression # PowerShell
```

A completion script only covers the name of the binary it was generated by. If you use the [other device families](#supported-device-families-multicall-binary), load the script of each variant as well:

```sh
. <(ovp9xx completion bash)
. <(o3cxxx completion bash)
```

See `ovp8xx completion <shell> --help` for further details, e.g. how to install the script system wide.

### CLI Usage

All variants offer the same commands, the examples below use `ovp8xx` but work the same way with `ovp9xx` and `o3cxxx`. Use `--help` on any command for the full list of options.

Select the device either by the `--ip` flag or by the environment variable of the variant:

```sh
ovp8xx --ip 192.168.0.69 get
export OVP9XX_IP=192.168.0.70
ovp9xx get
```

Without a link the variant can also be chosen by `OVP8XX_VARIANT`:

```sh
OVP8XX_VARIANT=o3cxxx ovp8xx get
```

Query the configuration, optionally limited by JSON pointers and pretty printed. See the [filter documentation](doc/filter.md) for transforming the result:

```sh
ovp8xx get --pretty
ovp8xx get -p /device/swVersion -p /ports/port0
```

Write a configuration back to the device, either as an argument or via stdin:

```sh
ovp8xx set '{"device":{"info":{"name":"my-device"}}}'
ovp8xx get -p /device/info | ovp8xx set
```

Persist the current configuration as the init configuration and read it back:

```sh
ovp8xx saveInit -p /device/info
ovp8xx getInit --pretty
```

Further commands:

```sh
ovp8xx getSchema -p /device --pretty        # JSON schema of the configuration
ovp8xx diagnostic getFiltered               # active diagnostic events
ovp8xx pcic --port 50010                    # receive PCIC result data
ovp8xx reboot                               # reboot the device
ovp8xx WaitForOnline --timeout 120          # wait until the device is reachable again
```

Update the firmware, the device has to be rebooted into the SWUpdate mode first:

```sh
ovp8xx reboot --swupdate
ovp8xx swupdate firmware.swu
ovp8xx swupdate restart                     # reboot back into productive mode
```

### API usage

Within in your Go project get the ovp8xx package first

```
go get github.com/graugans/go-ovp8xx/v2
```

The following example will query the software Version of your OVP8xx. Without `ovp8xx.WithHost(...)` the client connects to the default IP address `192.168.0.69` (`ovp8xx.DefaultHost`). The library does not read any environment variables. To resolve the host the same way as the CLI does (`OVP8XX_IP` and the variant specific variables such as `OVP9XX_IP`), use the `variant` package. All methods take a `context.Context`, which allows to cancel a request or limit its duration; `ovp8xx.WithTimeout(...)` sets a limit for every request of a client:

```go
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/graugans/go-ovp8xx/v2/pkg/ovp8xx"
	"github.com/graugans/go-ovp8xx/v2/pkg/variant"
)

func main() {
	o3r := ovp8xx.NewClient(
		ovp8xx.WithHost(variant.Default.DefaultHost(os.Getenv)),
	)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	config, err := o3r.Get(ctx, []string{"/device/swVersion"})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(config)
}
```

The result depends on your specific device:

```json
{
  "device": {
    "swVersion": {
      "euphrates": "1.32.1+8.e72bf7bb5",
      "firmware": "1.1.2-1335",
      "kernel": "4.9.140-l4t-r32.4+g8c7b68130d9a",
      "l4t": "r32.4.3",
      "schema": "v1.5.3",
      "tcu": "1.1.0"
    }
  }
}
```

## Testing

Please ensure you have the Git Large File Storage extension installed. Some of the tests require blobs which are handled as Large File Storage files. In case the files are not populated as expected this may help:

```sh
git lfs fetch --all
git lfs pull
```
