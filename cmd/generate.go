/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/sacloud/api-client-go/profile"
	"github.com/spf13/cobra"
)

type Resource struct {
	Id   string
	Data Service
}

type TrackedResource struct {
	ResourceName  string
	TerraformType string
	Options       []TfParameter
	ReturnValues  map[string]string
}

type Service interface {
	ServiceMapping(trackedResources *[]TrackedResource)
}

type TfParameter struct {
	Key   string
	Value any
}

type Config struct {
	Profile           string
	AccessToken       string
	AccessTokenSecret string
	Zone              string
}

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate Terraform template from your existiong SAKURA Cloud",
	Long:  `Generate Terraform template from your existiong SAKURA Cloud`,
	Run: func(cmd *cobra.Command, args []string) {

		config := Config{}
		if err := config.loadFromProfile(); err != nil {
			fmt.Println("failed to load profile:", err)
			os.Exit(1)
		}

		services, err := cmd.Flags().GetStringSlice("services")
		if err != nil {
			fmt.Println("failed to retrieve services:", err)
			os.Exit(1)
		}

		excludeServices, err := cmd.Flags().GetStringSlice("exclude-services")

		if err != nil {
			fmt.Println("failed to retrieve exclude services:", err)
			os.Exit(1)
		}

		sections, err := filterSections(services, excludeServices)

		if err != nil {
			fmt.Println("failed to filter section", err)
			os.Exit(1)
		}

		var resources []Resource

		for _, section := range sections {
			switch section {
			case "Disk":
				if err = updateDatatableDisk(&resources, &config); err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			case "Server":
				if err = updateDatatableServer(&resources, &config); err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			case "Switch":
				if err = updateDatatableSwitch(&resources, &config); err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			case "Internet":
				if err = updateDatatableInternet(&resources, &config); err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			}
		}

		searchFilter, err := cmd.Flags().GetString("filter")

		if err != nil {
			fmt.Println("failed to retrieving searchFilter:", err)
			os.Exit(1)
		}

		outputResources, err := filterResource(searchFilter, &resources)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		trackedResources := performMapping(outputResources)

		mappedOutputs := compileOutputs(trackedResources)
		output, err := cmd.Flags().GetString("output")

		if err != nil {
			fmt.Println("failed to retrieve output:", err)
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
	generateCmd.Flags().StringP("output", "o", "output.tf", "filename for Terraform output")
}

func getAllSections() []string {
	sections := []string{}
	sections = append(sections, "Server", "Disk", "Switch", "Internet")
	return sections
}

// Filtering the search service
func filterSections(services []string, excludeServices []string) ([]string, error) {
	sections := []string{}
	allSections := getAllSections()

	if len(services) > 0 && len(excludeServices) > 0 {
		return nil, fmt.Errorf("please do not use --exclude-services and --services simultaneously")
	} else if len(excludeServices) > 0 {
		for _, section := range allSections {
			if !slices.Contains(excludeServices, strings.ToLower(section)) {
				sections = append(sections, section)
			}
		}
	} else if len(services) > 0 {
		for _, section := range allSections {
			if slices.Contains(services, strings.ToLower(section)) {
				sections = append(sections, section)
			}
		}
	} else if len(services) == 0 && len(excludeServices) == 0 {
		sections = allSections
	}
	return sections, nil
}

func performMapping(outputResources []Resource) []TrackedResource {
	var trackedResources []TrackedResource
	for _, outputResource := range outputResources {
		outputResource.Data.ServiceMapping(&trackedResources)
	}

	return trackedResources
}

// Output Terraform template
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

// Output terraform resource
func outputMapTf(trackedResource TrackedResource, trackedResources []TrackedResource) string {

	var params string

	for _, option := range trackedResource.Options {
		switch v := option.Value.(type) {
		case map[string]string:
			optionValue := processTfParameter(4, option.Key, v, trackedResources)
			params += fmt.Sprintf(`
    %s %s`, option.Key, optionValue)
		default:
			optionValue := processTfParameter(4, option.Key, v, trackedResources)
			params += fmt.Sprintf(`
    %s = %s`, option.Key, optionValue)
		}

	}

	output := fmt.Sprintf(`
resource "%s" "%s" {%s
}`, trackedResource.TerraformType, trackedResource.ResourceName, params)

	return output
}

func processTfParameter(spacing int, k string, v any, trackedResources []TrackedResource) string {
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
			paramItems = append(paramItems, processTfParameter(spacing+4, k, param, trackedResources))
		}
		return "[" + strings.Join(paramItems, ",") + "]"
	case map[string]string:
		for key, value := range v {
			subValue := processTfParameter(spacing+4, key, value, trackedResources)
			paramItems = append(paramItems, key+" = "+subValue)
		}
		return `{
` + strings.Repeat(" ", spacing+4) + strings.Join(paramItems, `
`+strings.Repeat(" ", spacing+4)) + `
` + strings.Repeat(" ", spacing) + `}`

	default:
		return "" //[TODO]
	}
}

// Retrieve Existing Resources by Making a Request to the SAKURA Cloud API[TODO: Change to using library]
func callApi[T any](response *T, serviceName string, config *Config) error {
	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://secure.sakura.ad.jp/cloud/zone/"+config.Zone+"/api/cloud/1.1/"+serviceName+"/", nil)

	if err != nil {
		return err
	}

	req.SetBasicAuth(config.AccessToken, config.AccessTokenSecret)

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return err
	}
	return nil
}

// Filtering the Output Resources
func filterResource(searchFilter string, resources *[]Resource) ([]Resource, error) {
	var outputResources []Resource

	if searchFilter == "" {
		outputResources = append(outputResources, *resources...)
		return outputResources, nil
	}

	for _, resource := range *resources {
		jsonResBytes, err := json.Marshal(resource)

		if err != nil {
			return nil, err
		}

		jsonResString := string(jsonResBytes)

		if strings.Contains(searchFilter, ",") {

			for _, searchTerm := range strings.Split(searchFilter, ",") {
				if strings.Contains(jsonResString, searchTerm) {
					outputResources = append(outputResources, resource)
					break
				}
			}

		} else if strings.Contains(searchFilter, "&") {

			searchWords := strings.Split(searchFilter, "&")

			if isAllContains(jsonResString, searchWords) {
				outputResources = append(outputResources, resource)
			}
		} else {

			jsonResString := string(jsonResBytes)
			if strings.Contains(jsonResString, searchFilter) {
				outputResources = append(outputResources, resource)
			}
		}
	}
	return outputResources, nil
}

// Check if the Slice Contains an Element with the Specified String
func isAllContains(str string, slice []string) bool {
	for _, item := range slice {
		if !strings.Contains(str, item) {
			return false
		}
	}
	return true
}

// Load SAKURA Cloud Profile from usacloud config file
func (c *Config) loadFromProfile() error {

	pcv := &profile.ConfigValue{}
	if err := profile.Load(c.Profile, pcv); err != nil {
		return fmt.Errorf("loading profile %q is failed: %s", c.Profile, err)
	}

	if c.AccessToken == "" {
		c.AccessToken = pcv.AccessToken
	}
	if c.AccessTokenSecret == "" {
		c.AccessTokenSecret = pcv.AccessTokenSecret
	}
	if c.Zone == "" && pcv.Zone != "" {
		c.Zone = pcv.Zone
	}

	return nil
}
