/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"slices"

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
		allSections := getAllSections()
		services, err := cmd.Flags().GetStringSlice("services")
		if err != nil {
			fmt.Println("Error retrieving services:", err)
			return
		}

		sections := []section{}

		for _, section := range allSections {
			if (slices.Contains(services, section.service)){
				sections = append(sections, section)
			}
		}

		fmt.Println("Services:", services)
		fmt.Println("Sections:", sections)
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringSliceP("services", "s", []string{}, "list of services to include (can be comma separated (default: ALL))")
}

func getAllSections()[]section{
	sections := []section{}
	sections = append(sections, section{"server"}, section{"disk"}, section{"switch"})
	return sections
}