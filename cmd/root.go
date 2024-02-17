/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "henkan [global options] <command> <sub-command> [options]",
	Short: "henkan is CLI for generate Terraform templates from your existiong SAKURA Cloud resources",
	Long:  `CLI for generate terraform templates from your existiong SAKURA Cloud resources`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
