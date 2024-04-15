package ovp8xx

import (
	"encoding/json"
	"fmt"
	"slices"
)

// PCICInfo represents the information about a PCIC port.
type PCICInfo struct {
	Port            uint16   // Port is the port number.
	AvailableOutput []string // AvailableOutput is a list of available output options.
}

// PortInfo represents information about a port.
type PortInfo struct {
	Name  string   // Name of the port.
	State string   // State of the port.
	Type  string   // Type of the port.
	PCIC  PCICInfo // PCIC information for the port.
}

// Ports returns a slice of PortInfo structs, representing the available ports
// If there is an error retrieving the ports from the device or parsing the port info,
// an empty slice and the error are returned.
func (device *Client) Ports() ([]PortInfo, error) {
	emptyPorts := make([]PortInfo, 0)
	var err error

	info := struct {
		Ports map[string]struct {
			Data struct {
				AvailablePCICOutput []string `json:"availablePCICOutput"`
				PcicTCPPort         uint16   `json:"pcicTCPPort"`
			} `json:"data"`
			Info struct {
				Features struct {
					Type string `json:"type"`
				}
			} `json:"info"`
			State string `json:"state"`
		} `json:"ports"`
	}{}
	conf, err := device.Get([]string{"/ports"})
	if err != nil {
		return emptyPorts, fmt.Errorf("unable to get the ports from the device: %w", err)
	}
	err = json.Unmarshal([]byte(conf.String()), &info)
	if err != nil {
		return emptyPorts, fmt.Errorf("unable retrieve the port info from the input: %s; err: %w", conf.String(), err)
	}

	portNames := make([]string, 0)
	for key := range info.Ports {
		portNames = append(portNames, key)
	}
	slices.Sort(portNames)
	ports := make([]PortInfo, len(portNames))
	for idx, key := range portNames {
		ports[idx] = PortInfo{
			Name:  key,
			State: info.Ports[key].State,
			Type:  info.Ports[key].Info.Features.Type,
			PCIC: PCICInfo{
				Port:            info.Ports[key].Data.PcicTCPPort,
				AvailableOutput: info.Ports[key].Data.AvailablePCICOutput,
			},
		}
	}

	return ports, err
}
