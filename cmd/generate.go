/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

type section struct {
	service string
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
		fmt.Println("generate called")
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

		fmt.Println("Services:", services)
		fmt.Println("Sections:", sections)

		updateDatatableServer()
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringSliceP("services", "s", []string{}, "list of services to include (can be comma separated (default: ALL))")
	generateCmd.Flags().StringSliceP("exclude-services", "e", []string{}, "list of services to exclude (can be comma separated)")
}

func getAllSections()[]section{
	sections := []section{}
	sections = append(sections, section{"Server"}, section{"Disk"}, section{"Switch"})
	return sections
}

func filterSections(services []string, excludeServices []string)([]section, error){
	sections := []section{}
	allSections := getAllSections()

	if len(services) >0 && len(excludeServices) >0 {
		return nil, errors.New("Please do not use --exclude-services and --services simultaneously")
	} else if len(excludeServices) >0 {
		for _, section := range allSections {
			if (!slices.Contains(excludeServices, strings.ToLower(section.service))){
				sections = append(sections, section)
			}
		}
	} else if len(services) >0 {
		for _, section := range allSections {
			if (slices.Contains(services, strings.ToLower(section.service))){
				sections = append(sections, section)
			}
		}
	} else if len(services) ==0 && len(excludeServices)==0 {
		sections = allSections
	}
	return sections, nil
}

func updateDatatableServer(){

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

	fmt.Println(string(body))
}