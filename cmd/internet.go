/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
)

type InternetResponse struct {
	Internet []Internet `json:"Internet"`
}

type Internet struct {
	ID             string   `json:"ID"`
	Name           string   `json:"Name"`
	NetworkMaskLen int      `json:"NetworkMaskLen"`
	BandWidthMbps  int      `json:"BandWidthMbps"`
	Description    string   `json:"Description"`
	Tags           []string `json:"Tags"`
}

func updateDatatableInternet(resources *[]Resource, config *Config) error {

	var internetResponse InternetResponse

	if err := callApi(&internetResponse, "internet", config); err != nil {
		return fmt.Errorf("failed to call sakuracloud api: %w", err)
	}

	for _, internet := range internetResponse.Internet {
		*resources = append(*resources, Resource{Id: internet.Name, Data: internet})
	}
	return nil
}

func (i Internet) ServiceMapping(trackedResources *[]TrackedResource) {
	options := make(map[string]any)

	options["name"] = i.Name
	options["description"] = i.Description
	options["band_width"] = i.BandWidthMbps
	options["netmask"] = i.NetworkMaskLen
	options["tags"] = i.Tags

	returnValues := make(map[string]string)
	returnValues["id"] = i.ID

	*trackedResources = append(*trackedResources, TrackedResource{ResourceName: i.Name, TerraformType: "sakuracloud_internet", Options: options, ReturnValues: returnValues})
}
