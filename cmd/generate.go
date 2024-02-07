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

	"github.com/spf13/cobra"
)

type Section struct {
	service string
}

type Resource struct {
	Id   string
	Type string
	Data Service
}

type OutputResource struct {
	Id   string
	Type string
	Data Service
}

type TrackedResource struct {
	ResourceName  string
	TerraformType string
	Options       map[string]any
	ReturnValues  map[string]string
}

type Service interface {
	ServiceMapping(trackedResources *[]TrackedResource)
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

		searchFilter, err := cmd.Flags().GetString("filter")

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		for _, resource := range resources {
			if searchFilter == "" {
				outputResources = append(outputResources, OutputResource{Id: resource.Id, Type: resource.Type, Data: resource.Data})
			} else if strings.Contains(searchFilter, ",") {
				jsonResBytes, err := json.Marshal(resource)

				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}

				jsonResString := string(jsonResBytes)

				for _, searchTerm := range strings.Split(searchFilter, ",") {
					if strings.Contains(jsonResString, searchTerm) {
						outputResources = append(outputResources, OutputResource{Id: resource.Id, Type: resource.Type, Data: resource.Data})
						break
					}
				}

			} else if strings.Contains(searchFilter, "&") {
				jsonResBytes, err := json.Marshal(resource)

				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}

				jsonResString := string(jsonResBytes)

				searchWords := strings.Split(searchFilter, "&")

				if isAllContains(jsonResString, searchWords) {
					outputResources = append(outputResources, OutputResource{Id: resource.Id, Type: resource.Type, Data: resource.Data})
				}
			} else {
				jsonResBytes, err := json.Marshal(resource)

				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}

				jsonResString := string(jsonResBytes)
				if strings.Contains(jsonResString, searchFilter) {
					outputResources = append(outputResources, OutputResource{Id: resource.Id, Type: resource.Type, Data: resource.Data})
				}
			}
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
	generateCmd.Flags().StringP("filter", "f", "", "search filter for discovered resources (can be comma separated)")
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

func performMapping(outputResources []OutputResource) []TrackedResource {
	var trackedResources []TrackedResource
	for _, outputResource := range outputResources {
		outputResource.Data.ServiceMapping(&trackedResources)
	}

	return trackedResources
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
}`, trackedResource.TerraformType, trackedResource.ResourceName, params)

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
						return trackedResource.TerraformType + "." + trackedResource.ResourceName + "." + key
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

func callApi[T any](response *T, serviceName string) error {
	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://secure.sakura.ad.jp/cloud/zone/is1a/api/cloud/1.1/"+serviceName+"/", nil)

	if err != nil {
		fmt.Println(err)
		return err
	}

	access_token := os.Getenv("SAKURACLOUD_ACCESS_TOKEN")
	access_token_secret := os.Getenv("SAKURACLOUD_ACCESS_TOKEN_SECRET")
	req.SetBasicAuth(access_token, access_token_secret)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return err
	}

	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func isAllContains(str string, slice []string) bool {
	for _, item := range slice {
		if !strings.Contains(str, item) {

			return false
		}
	}
	return true
}
