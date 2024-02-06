/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/RyoMasumura1201/henkan/service"
	"github.com/spf13/cobra"
)

type Section struct {
	service string
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
	ReturnValues   map[string]string
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
			os.Exit(1)
		}

		excludeServices, err := cmd.Flags().GetStringSlice("exclude-services")

		if err != nil {
			fmt.Println("Error retrieving exclude services:", err)
			os.Exit(1)
		}

		sections, err := filterSections(services, excludeServices)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println("Sections:", sections)

		var resources []Resource

		for _, section := range sections {
			if section.service == "Disk" {
				if err = updateDatatableDisk(&resources); err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			} else if section.service == "Server" {
				if err = updateDatatableServer(&resources); err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			}
		}

		var outputResources []OutputResource

		// [TODO] search filter
		for _, resource := range resources {
			outputResources = append(outputResources, OutputResource{Id: resource.Id, Type: resource.Type, Data: resource.Data})
		}

		trackedResources := performMapping(outputResources)

		mappedOutputs := compileOutputs(trackedResources)
		output, err := cmd.Flags().GetString("output")

		if err != nil {
			fmt.Println("Error retrieving output:", err)
			os.Exit(1)
		}
		if err = os.WriteFile(output, []byte(mappedOutputs), 0666); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringSliceP("services", "s", []string{}, "list of services to include (can be comma separated (default: ALL))")
	generateCmd.Flags().StringSliceP("exclude-services", "e", []string{}, "list of services to exclude (can be comma separated)")
	generateCmd.Flags().StringP("output", "o", "example.tf", "filename for Terraform output")
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

func updateDatatableServer(resources *[]Resource) error {

	serverResponse, err := callApi[service.ServerResponse]("server")

	if err != nil {
		fmt.Println(err)
		return err
	}

	for _, server := range serverResponse.Servers {
		*resources = append(*resources, Resource{Id: server.Name, Type: "server", Name: server.Name, Data: server})
	}

	return nil
}

func updateDatatableDisk(resources *[]Resource) error {

	diskResponse, err := callApi[service.DiskResponse]("disk")

	if err != nil {
		fmt.Println(err)
		return err
	}
	for _, disk := range diskResponse.Disks {
		*resources = append(*resources, Resource{Id: disk.Name, Type: "disk", Name: disk.Name, Data: disk})
	}
	return nil
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
		server, ok := outputResource.Data.(service.Server)
		if !ok {
			panic("failed to assertion")
		}

		options["name"] = server.Name
		options["core"] = server.ServerPlan.CPU
		options["memory"] = server.ServerPlan.MemoryMB / 1024
		var diskIds []string
		for _, disk := range server.Disks {
			diskIds = append(diskIds, disk.Id)
		}
		options["disks"] = diskIds

		networkInterface := make(map[string]string)
		networkInterface["upstream"] = server.Interfaces[0].Switch.Scope

		options["network_interface"] = networkInterface

		diskEditParameter := make(map[string]string)
		diskEditParameter["hostname"] = server.HostName

		options["disk_edit_parameter"] = diskEditParameter

		options["tags"] = server.Tags
		options["description"] = server.Description

		returnValues := make(map[string]string)
		returnValues["id"] = server.ID

		*trackedResources = append(*trackedResources, TrackedResource{OutputResource: outputResource, Service: "server", TerraformType: "sakuracloud_server", Options: options, ReturnValues: returnValues})
	} else if outputResource.Type == "disk" {
		options := make(map[string]any)
		disk, ok := outputResource.Data.(service.Disk)
		if !ok {
			panic("failed to assertion")
		}

		options["name"] = disk.Name
		options["connector"] = disk.Connection
		options["size"] = disk.SizeMB / 1024
		options["plan"] = strings.ToLower(strings.Replace(disk.Plan.Name, "プラン", "", -1))

		returnValues := make(map[string]string)
		returnValues["id"] = disk.ID

		*trackedResources = append(*trackedResources, TrackedResource{OutputResource: outputResource, Service: "disk", TerraformType: "sakuracloud_disk", Options: options, ReturnValues: returnValues})
	}
}

func compileOutputs(trackedResources []TrackedResource) string {
	compiled := `terraform {
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
		compiled += outputMapTf(trackedResource, trackedResources)
	}

	return compiled
}

func outputMapTf(trackedResource TrackedResource, trackedResources []TrackedResource) string {

	var params string

	for k, v := range trackedResource.Options {
		switch v := v.(type) {
		case map[string]string:
			optionValue := processTfParameter(k, v, trackedResources)
			params += fmt.Sprintf(`
    %s %s`, k, optionValue)
		default:
			optionValue := processTfParameter(k, v, trackedResources)
			params += fmt.Sprintf(`
    %s = %s`, k, optionValue)
		}

	}

	output := fmt.Sprintf(`
resource "%s" "%s" {%s
}`, trackedResource.TerraformType, trackedResource.OutputResource.Id, params)

	return output
}

func processTfParameter(k string, v any, trackedResources []TrackedResource) string {
	var paramItems []string
	switch v := v.(type) {
	case string:
		for _, trackedResource := range trackedResources {
			if trackedResource.ReturnValues != nil {
				for key, value := range trackedResource.ReturnValues {
					if value == v {
						return trackedResource.TerraformType + "." + trackedResource.OutputResource.Id + "." + key
					}
				}
			}

		}
		return "\"" + v + "\""
	case int:
		return strconv.Itoa(v)
	case []string:
		for _, param := range v {
			paramItems = append(paramItems, processTfParameter(k, param, trackedResources))
		}
		return "[" + strings.Join(paramItems, ",") + "]"
	case map[string]string:
		for key, value := range v {
			subValue := processTfParameter(key, value, trackedResources)
			paramItems = append(paramItems, key+" = "+subValue)
		}
		return `{
        ` + strings.Join(paramItems, `
        `) + `
    }`

	default:
		return "" //[TODO]
	}
}

func callApi[T any](resourceName string) (*T, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://secure.sakura.ad.jp/cloud/zone/is1a/api/cloud/1.1/"+resourceName+"/", nil)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	access_token := os.Getenv("SAKURACLOUD_ACCESS_TOKEN")
	access_token_secret := os.Getenv("SAKURACLOUD_ACCESS_TOKEN_SECRET")
	req.SetBasicAuth(access_token, access_token_secret)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var response T
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Println(err)
		return nil, err
	}
	return &response, nil
}
