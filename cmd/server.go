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
	ID    string `json:"ID"`
	Scope string `json:"Scope"`
}

func updateDatatableServer(resources *[]Resource, config *Config) error {

	var serverResponse ServerResponse

	if err := callApi(&serverResponse, "server", config); err != nil {
		return fmt.Errorf("failed to call sakuracloud api: %w", err)
	}

	for _, server := range serverResponse.Servers {
		*resources = append(*resources, Resource{Id: server.Name, Data: server})
	}

	return nil
}

func (s Server) ServiceMapping(trackedResources *[]TrackedResource) {
	options := []TfParameter{}
	options = append(options, TfParameter{"name", s.Name})
	options = append(options, TfParameter{"core", s.ServerPlan.CPU})
	options = append(options, TfParameter{"memory", s.ServerPlan.MemoryMB / 1024})

	networkInterface := make(map[string]string)
	switch s.Interfaces[0].Switch.Scope {
	case "shared":
		networkInterface["upstream"] = s.Interfaces[0].Switch.Scope
	case "user":
		networkInterface["upstream"] = s.Interfaces[0].Switch.ID
	}
	options = append(options, TfParameter{"network_interface", networkInterface})

	var diskIds []string
	for _, disk := range s.Disks {
		diskIds = append(diskIds, disk.Id)
	}
	options = append(options, TfParameter{"disks", diskIds})

	diskEditParameter := make(map[string]string)
	diskEditParameter["hostname"] = s.HostName
	options = append(options, TfParameter{"disk_edit_parameter", diskEditParameter})

	options = append(options, TfParameter{"description", s.Description})
	options = append(options, TfParameter{"tags", s.Tags})

	returnValues := make(map[string]string)
	returnValues["id"] = s.ID

	*trackedResources = append(*trackedResources, TrackedResource{ResourceName: s.Name, TerraformType: "sakuracloud_server", Options: options, ReturnValues: returnValues})
}
