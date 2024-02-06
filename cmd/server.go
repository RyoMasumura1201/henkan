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

	serverResponse, err := callApi[ServerResponse]("server")

	if err != nil {
		fmt.Println(err)
		return err
	}

	for _, server := range serverResponse.Servers {
		*resources = append(*resources, Resource{Id: server.Name, Type: "server", Name: server.Name, Data: server})
	}

	return nil
}
