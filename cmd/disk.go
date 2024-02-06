/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

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
