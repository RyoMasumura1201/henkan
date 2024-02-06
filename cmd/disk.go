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
	ID         string   `json:"ID"`
	Name       string   `json:"Name"`
	SizeMB     int      `json:"SizeMB"`
	Connection string   `json:"Connection"`
	Plan       DiskPlan `json:"Plan"`
}

type DiskPlan struct {
	Name string `json:"Name"`
}

func updateDatatableDisk(resources *[]Resource) error {

	var diskResponse DiskResponse

	if err := callApi(&diskResponse, "disk"); err != nil {
		fmt.Println(err)
		return err
	}

	for _, disk := range diskResponse.Disks {
		*resources = append(*resources, Resource{Id: disk.Name, Type: "disk", Name: disk.Name, Data: disk})
	}
	return nil
}

func (d Disk) ServiceMapping(trackedResources *[]TrackedResource) {
	options := make(map[string]any)

	options["name"] = d.Name
	options["connector"] = d.Connection
	options["size"] = d.SizeMB / 1024
	options["plan"] = strings.ToLower(strings.Replace(d.Plan.Name, "プラン", "", -1))

	returnValues := make(map[string]string)
	returnValues["id"] = d.ID

	*trackedResources = append(*trackedResources, TrackedResource{ResourceName: d.Name, TerraformType: "sakuracloud_disk", Options: options, ReturnValues: returnValues})
}
