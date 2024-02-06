/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import "fmt"

type ServerResponse struct {
	Servers []Server `json:"Servers"`
}

type Server struct {
	ID          string            `json:"ID"`
	Name        string            `json:"Name"`
	HostName    string            `json:"HostName"`
	ServerPlan  ServerPlan        `json:"ServerPlan"`
	Disks       []ServerDisk      `json:"Disks"`
	Interfaces  []ServerInterface `json:"Interfaces"`
	Tags        []string          `json:"Tags"`
	Description string            `json:"Description"`
}

type ServerPlan struct {
	CPU      int `json:"CPU"`
	MemoryMB int `json:"MemoryMB"`
}

type ServerDisk struct {
	Id string `json:"ID"`
}

type ServerInterface struct {
	Switch ServerInterfaceSwitch `json:"Switch"`
}

type ServerInterfaceSwitch struct {
	Scope string `json:"Scope"`
}

func updateDatatableServer(resources *[]Resource) error {

	var serverResponse ServerResponse

	if err := callApi(&serverResponse, "server"); err != nil {
		fmt.Println(err)
		return err
	}

	for _, server := range serverResponse.Servers {
		*resources = append(*resources, Resource{Id: server.Name, Type: "server", Name: server.Name, Data: server})
	}

	return nil
}

func (s Server) ServiceMapping(trackedResources *[]TrackedResource) {
	options := make(map[string]any)

	options["name"] = s.Name
	options["core"] = s.ServerPlan.CPU
	options["memory"] = s.ServerPlan.MemoryMB / 1024
	var diskIds []string
	for _, disk := range s.Disks {
		diskIds = append(diskIds, disk.Id)
	}
	options["disks"] = diskIds

	networkInterface := make(map[string]string)
	networkInterface["upstream"] = s.Interfaces[0].Switch.Scope

	options["network_interface"] = networkInterface

	diskEditParameter := make(map[string]string)
	diskEditParameter["hostname"] = s.HostName

	options["disk_edit_parameter"] = diskEditParameter

	options["tags"] = s.Tags
	options["description"] = s.Description

	returnValues := make(map[string]string)
	returnValues["id"] = s.ID

	*trackedResources = append(*trackedResources, TrackedResource{ResourceName: s.Name, TerraformType: "sakuracloud_server", Options: options, ReturnValues: returnValues})
}
