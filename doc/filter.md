# Using Go text/template for Advanced Data Filtering

This document demonstrates the usage of the Go text/template package to filter and format the output of the `ovp8xx get` command for manipulating the result data. There are two custom functions created to enhance the filtering: `prefix` and `toJSON`.

## The `prefix` Function

The `prefix` function is designed to solve the problem of trailing commas, which are, for example, not allowed in JSON. The idea is to prefix each line with a comma, except for the very first one. This function is particularly useful when generating JSON output dynamically, where the number of elements may vary.

Here's how it's used in the command:

```sh
ovp8xx get  --format '{ "ports": { {{$p := prefix ", "}}{{- range $key, $val := .ports -}}{{call $p}}"{{- $key -}}":{"state": "{{ .state }}"}{{- end }} } }'
```

In this command, `{{$p := prefix ", "}}` initializes the prefix function with a comma and a space. Then, `{{call $p}}` is used to insert the prefix before each port entry. The prefix function ensures that no comma is inserted before the first entry.

The result is a JSON object where each line, except the first, is prefixed with a comma:

```sh
{
  "ports": {
    "port0": {
      "state": "RUN"
    },
    "port1": {
      "state": "RUN"
    },
    "port2": {
      "state": "RUN"
    },
    "port3": {
      "state": "RUN"
    },
    "port6": {
      "state": "RUN"
    }
  }
}
```

## Print all ports with some details

The following command retrieves all port data and formats it for easy reading:

```sh
ovp8xx get  --format '{{$p := prefix "\n"}}{{ range $port, $details := .ports }}{{call $p}}[{{ $port }}] state: {{ $details.state }},{{print "\t"}}type: {{ $details.info.features.type }},{{print "\t"}}PCIC Port: {{ $details.data.pcicTCPPort }}{{ end }}'
```

The result will look like:

```
[port0] state: RUN,     type: 3D,       PCIC Port: 50010
[port1] state: RUN,     type: 3D,       PCIC Port: 50011
[port2] state: RUN,     type: 2D,       PCIC Port: 50012
[port3] state: RUN,     type: 2D,       PCIC Port: 50013
[port6] state: RUN,     type: IMU,      PCIC Port: 50016
```

## Modify the state of the ports

Now we are taking the example from the [`prefix`](#the-prefix-function) function and enhance it to also modify the state of all ports (except port6) to CONF with the following command:

```sh
ovp8xx get  --format '{{ $state := "CONF" }}{ "ports": { {{$p := prefix ", "}}{{- range $key, $val := .ports -}}{{ if ne $key "port6" }}{{call $p}}"{{- $key -}}":{"state": "{{ $state }}"}{{end}}{{- end }} } }'
```

The output will be a JSON object that can be directly piped into the set command:

```sh
{
    "ports": {
        "port0": {
            "state": "CONF"
        },
        "port1": {
            "state": "CONF"
        },
        "port2": {
            "state": "CONF"
        },
        "port3": {
            "state": "CONF"
        }
    }
}
```

To directly apply those changes to the device, use the following command:

```sh
ovp8xx get  --format '{{ $state := "CONF" }}{ "ports": { {{$p := prefix ", "}}{{- range $key, $val := .ports -}}{{ if ne $key "port6" }}{{call $p}}"{{- $key -}}":{"state": "{{ $state }}"}{{end}}{{- end }} } }'  | ovp8xx set
```
