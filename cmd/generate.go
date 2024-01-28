/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

type Section struct {
	service string
}

type ServerResponse struct {
	From    int      `json:"From"`
	Count   int      `json:"Count"`
	Total   int      `json:"Total"`
	Servers []Server `json:"Servers"`
	IsOK    bool     `json:"is_ok"`
}

type Server struct {
	Name       string     `json:"Name"`
	ServerPlan ServerPlan `json:"ServerPlan"`
}

type ServerPlan struct {
	CPU      int `json:"CPU"`
	MemoryMB int `json:"MemoryMB"`
}

type Resource struct {
	Id   string
	Type string
	Name string
	Data any
}

type OutputResource struct {
	Id   string
	Type string
	Data any
}

type TrackedResource struct {
	OutputResource OutputResource
	Service        string
	TerraformType  string
	Options        map[string]any
}

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		services, err := cmd.Flags().GetStringSlice("services")
		if err != nil {
			fmt.Println("Error retrieving services:", err)
			return
		}

		excludeServices, err := cmd.Flags().GetStringSlice("exclude-services")

		if err != nil {
			fmt.Println("Error retrieving exclude services:", err)
			return
		}

		sections, err := filterSections(services, excludeServices)

		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("Sections:", sections)

		var resources []Resource

		updateDatatableServer(&resources)

		var outputResources []OutputResource

		// [TODO] search filter
		for _, resource := range resources {
			outputResources = append(outputResources, OutputResource{Id: resource.Id, Type: resource.Type, Data: resource.Data})
		}

		trackedResources := performMapping(outputResources)

		compileOutputs(trackedResources)
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringSliceP("services", "s", []string{}, "list of services to include (can be comma separated (default: ALL))")
	generateCmd.Flags().StringSliceP("exclude-services", "e", []string{}, "list of services to exclude (can be comma separated)")
}

func getAllSections() []Section {
	sections := []Section{}
	sections = append(sections, Section{"Server"}, Section{"Disk"}, Section{"Switch"})
	return sections
}

func filterSections(services []string, excludeServices []string) ([]Section, error) {
	sections := []Section{}
	allSections := getAllSections()

	if len(services) > 0 && len(excludeServices) > 0 {
		return nil, errors.New("Please do not use --exclude-services and --services simultaneously")
	} else if len(excludeServices) > 0 {
		for _, section := range allSections {
			if !slices.Contains(excludeServices, strings.ToLower(section.service)) {
				sections = append(sections, section)
			}
		}
	} else if len(services) > 0 {
		for _, section := range allSections {
			if slices.Contains(services, strings.ToLower(section.service)) {
				sections = append(sections, section)
			}
		}
	} else if len(services) == 0 && len(excludeServices) == 0 {
		sections = allSections
	}
	return sections, nil
}

func updateDatatableServer(resources *[]Resource) {

	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://secure.sakura.ad.jp/cloud/zone/is1a/api/cloud/1.1/server/", nil)

	if err != nil {
		fmt.Println(err)
		return
	}

	access_token := os.Getenv("SAKURACLOUD_ACCESS_TOKEN")
	access_token_secret := os.Getenv("SAKURACLOUD_ACCESS_TOKEN_SECRET")
	req.SetBasicAuth(access_token, access_token_secret)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	var serverResponse ServerResponse
	if err := json.Unmarshal(body, &serverResponse); err != nil {
		fmt.Println(err)
		return
	}

	for _, server := range serverResponse.Servers {
		*resources = append(*resources, Resource{Id: server.Name, Type: "server", Name: server.Name, Data: server})
	}
}

func performMapping(outputResources []OutputResource) []TrackedResource {
	var trackedResources []TrackedResource
	for _, outputResource := range outputResources {
		serviceMapping(outputResource, &trackedResources)
	}

	return trackedResources
}

func serviceMapping(outputResource OutputResource, trackedResources *[]TrackedResource) {
	if outputResource.Type == "server" {
		options := make(map[string]any)
		server, ok := outputResource.Data.(Server)
		if !ok {
			panic("failed to assertion")
		}

		options["name"] = server.Name
		options["core"] = server.ServerPlan.CPU
		options["memory"] = server.ServerPlan.MemoryMB / 1024

		*trackedResources = append(*trackedResources, TrackedResource{OutputResource: outputResource, Service: "server", TerraformType: "sakuracloud_server", Options: options})
	}
}

func compileOutputs(trackedResources []TrackedResource) {
	compiled := `
terraform {
    required_providers {
        sakuracloud = {
            source  = "sacloud/sakuracloud"
            version = "2.25.0"
        }
    }
}
provider "sakuracloud" {
    profile = "default"
}
`

	for _, trackedResource := range trackedResources {
		compiled += outputMapTf(trackedResource)
	}

	fmt.Println(compiled)
}

func outputMapTf(trackedResource TrackedResource) string {

	var params string

	fmt.Println(trackedResource.Options)

	for k, v := range trackedResource.Options {
		processTfParameter(k, v)
		switch v := v.(type) {
		case string:
			v = "\"" + v + "\""
			params += fmt.Sprintf(`
    %s = %s`, k, v)
		case int:
			params += fmt.Sprintf(`
    %s = %d`, k, v)
		}

	}

	output := fmt.Sprintf(`
resource %s "%s" {%s
}`, trackedResource.TerraformType, trackedResource.OutputResource.Id, params)

	return output
}

// [TODO]
func processTfParameter(k string, v any) {

}
