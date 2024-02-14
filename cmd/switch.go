/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
)

type SwitchResponse struct {
	Switches []Switch `json:"Switches"`
}

type Switch struct {
	ID          string   `json:"ID"`
	Name        string   `json:"Name"`
	Description string   `json:"Description"`
	Tags        []string `json:"Tags"`
}

func updateDatatableSwitch(resources *[]Resource, config *Config) error {

	var switchResponse SwitchResponse

	if err := callApi(&switchResponse, "switch", config); err != nil {
		return fmt.Errorf("failed to call sakuracloud api: %w", err)
	}

	for _, switch_resource := range switchResponse.Switches {
		*resources = append(*resources, Resource{Id: switch_resource.Name, Data: switch_resource})
	}
	return nil
}

func (s Switch) ServiceMapping(trackedResources *[]TrackedResource) {
	options := []TfParameter{}

	options = append(options, TfParameter{"name", s.Name})
	options = append(options, TfParameter{"description", s.Description})
	options = append(options, TfParameter{"tags", s.Tags})

	returnValues := make(map[string]string)
	returnValues["id"] = s.ID

	*trackedResources = append(*trackedResources, TrackedResource{ResourceName: s.Name, TerraformType: "sakuracloud_switch", Options: options, ReturnValues: returnValues})
}
