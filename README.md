

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

### CLI  Installation

#### Pre Build Binaries

The recommended and easiest way is to download the pre-build binary from the [GitHub Release page](https://github.com/graugans/go-ovp8xx/releases).

⚠️ The Windows binary maybe flagged by a Virus Scanner, please also read the note from the [Go Team](https://go.dev/doc/faq#virus) ⚠️

#### Go get

If you have a decent Go version installed

```sh
go install github.com/graugans/go-ovp8xx/v2/cmd/ovp8xx@latest
```

### API usage

Within in your Go project get the ovp8xx package first

```
go get github.com/graugans/go-ovp8xx/v2
```

The following example will query the software Version of your OVP8xx. This assumes that either the OVP8xx is using the default IP address of `192.168.0.69` or the environment variable `OVP8XX_IP` is set. In case you want to set the IP in code please use `ovp8xx.NewClient(ovp8xx.WithHost("192.168.0.69"))` to construct the client.

```go
package main

import (
	"fmt"

	"github.com/graugans/go-ovp8xx/v2/pkg/ovp8xx"
)

func main() {
	o3r := ovp8xx.NewClient()
	config, err := o3r.Get([]string{"/device/swVersion"})
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
