/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"slices"
	"testing"
)

func TestFilterSections(t *testing.T) {
	tests := []struct {
		name            string
		services        []string
		excludeServices []string
		want            []Section
	}{
		{name: "include 1 service", services: []string{"server"}, excludeServices: []string{}, want: []Section{{service: "Server"}}},
		{name: "include multiple services", services: []string{"server", "disk"}, excludeServices: []string{}, want: []Section{{service: "Server"}, {service: "Disk"}}},
		{name: "exclude 1 service", services: []string{}, excludeServices: []string{"server"}, want: []Section{{service: "Disk"}, {service: "Switch"}}},
		{name: "exclude multiple services", services: []string{}, excludeServices: []string{"server", "disk"}, want: []Section{{service: "Switch"}}},
		{name: "specify no service", services: []string{}, excludeServices: []string{}, want: []Section{{service: "Server"}, {service: "Disk"}, {service: "Switch"}}},
	}

	for _, tc := range tests {
		got, err := filterSections(tc.services, tc.excludeServices)
		if err != nil {
			t.Fatalf("%s: expected: %v, got: %v,", tc.name, tc.want, got)
		}
		if !slices.Equal(tc.want, got) {
			t.Fatalf("%s: expected: %v, got: %v,", tc.name, tc.want, got)
		}
	}

}
