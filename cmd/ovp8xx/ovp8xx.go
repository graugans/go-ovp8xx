package main

import (
	"fmt"

	"github.com/graugans/go-ovp8xx/pkg/config"
)

func main() {
	o3r := config.OVP8xx{
		Host: "172.25.125.66",
	}
	query := []string{"/device"}
	result, err := o3r.Get(query)
	if err != nil {
		fmt.Printf("Error occured: %s", err)
	} else {
		fmt.Printf("%s\n", result)
	}

}
