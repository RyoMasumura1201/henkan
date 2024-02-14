/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"strings"
)

type DiskResponse struct {
	Disks []Disk `json:"Disks"`
}

type Disk struct {
	ID          string   `json:"ID"`
	Name        string   `json:"Name"`
	SizeMB      int      `json:"SizeMB"`
	Connection  string   `json:"Connection"`
	Plan        DiskPlan `json:"Plan"`
	Description string   `json:"Description"`
	Tags        []string `json:"Tags"`
}

type DiskPlan struct {
	Name string `json:"Name"`
}

func updateDatatableDisk(resources *[]Resource, config *Config) error {

	var diskResponse DiskResponse

	if err := callApi(&diskResponse, "disk", config); err != nil {
		return fmt.Errorf("failed to call sakuracloud api: %w", err)
	}

	for _, disk := range diskResponse.Disks {
		*resources = append(*resources, Resource{Id: disk.Name, Data: disk})
	}
	return nil
}

func (d Disk) ServiceMapping(trackedResources *[]TrackedResource) {
	options := []TfParameter{}

	options = append(options, TfParameter{"name", d.Name})
	options = append(options, TfParameter{"plan", strings.ToLower(strings.Replace(d.Plan.Name, "プラン", "", -1))})
	options = append(options, TfParameter{"size", d.SizeMB / 1024})
	options = append(options, TfParameter{"connector", d.Connection})
	options = append(options, TfParameter{"description", d.Description})
	options = append(options, TfParameter{"tags", d.Tags})

	returnValues := make(map[string]string)
	returnValues["id"] = d.ID

	*trackedResources = append(*trackedResources, TrackedResource{ResourceName: d.Name, TerraformType: "sakuracloud_disk", Options: options, ReturnValues: returnValues})
}
