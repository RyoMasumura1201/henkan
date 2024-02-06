/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import "fmt"

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

	err := callApi(&diskResponse, "disk")

	if err != nil {
		fmt.Println(err)
		return err
	}
	for _, disk := range diskResponse.Disks {
		*resources = append(*resources, Resource{Id: disk.Name, Type: "disk", Name: disk.Name, Data: disk})
	}
	return nil
}
