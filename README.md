

# Go client for the ifm OVP8xx series of devices

A GO module and cli to access the ifm OVP8xx series of devices.

[![GitHub go.mod Go version of a Go module](https://img.shields.io/github/go-mod/go-version/gomods/athens.svg)](https://github.com/gomods/athens)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
![example workflow](https://github.com/graugans/go-ovp8xx/actions/workflows/go.yml/badge.svg)
[![codecov](https://codecov.io/gh/graugans/go-ovp8xx/graph/badge.svg?token=BU6UPYCUPI)](https://codecov.io/gh/graugans/go-ovp8xx)

## Project status

This project is still a work in progress and will suffer from breaking API changes. Please be warned. In case you have any suggestions or want to contribute please feel free to open an issue or pull request.

## CLI  Installation

### Pre Build Binaries

The recommended and easiest way is to download the pre-build binary from the [GitHub Release page](https://github.com/graugans/go-ovp8xx/releases).


### Go get

If you have a decent Go version installed

```sh
go install github.com/graugans/go-ovp8xx/cmd/ovp8xx@latest
```

## API usage

Within in your Go project get the ovp8xx package first

```
go get github.com/graugans/go-ovp8xx
```

The following example will query the software Version of your OVP8xx. This assumes that either the OVP8xx is using the default IP address of `192.168.0.69` or the environment variable `OVP8XX_IP` is set. In case you want to set the IP in code please use `ovp8xx.NewClient(ovp8xx.WithHost("192.168.0.69"))` to construct the client.

```go
package main

import (
	"fmt"

	"github.com/graugans/go-ovp8xx/pkg/ovp8xx"
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