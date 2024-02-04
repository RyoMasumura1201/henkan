/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package service

type ServerResponse struct {
	Servers []Server `json:"Servers"`
}

type Server struct {
	Name       string            `json:"Name"`
	HostName   string            `json:"HostName"`
	ServerPlan ServerPlan        `json:"ServerPlan"`
	Disks      []ServerDisk      `json:"Disks"`
	Interfaces []ServerInterface `json:"Interfaces"`
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
