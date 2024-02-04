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
		expectErr       bool
	}{
		{name: "include 1 service", services: []string{"server"}, excludeServices: []string{}, want: []Section{{service: "Server"}}, expectErr: false},
		{name: "include multiple services", services: []string{"server", "disk"}, excludeServices: []string{}, want: []Section{{service: "Server"}, {service: "Disk"}}, expectErr: false},
		{name: "exclude 1 service", services: []string{}, excludeServices: []string{"server"}, want: []Section{{service: "Disk"}, {service: "Switch"}}, expectErr: false},
		{name: "exclude multiple services", services: []string{}, excludeServices: []string{"server", "disk"}, want: []Section{{service: "Switch"}}, expectErr: false},
		{name: "specify no service", services: []string{}, excludeServices: []string{}, want: []Section{{service: "Server"}, {service: "Disk"}, {service: "Switch"}}, expectErr: false},
		{name: "specify service and exclude service", services: []string{"server"}, excludeServices: []string{"disk"}, want: []Section{}, expectErr: true},
	}

	for _, tc := range tests {
		got, err := filterSections(tc.services, tc.excludeServices)
		if tc.expectErr && err == nil {
			t.Fatalf("%s: expected: error, got: no error,", tc.name)
		}
		if !tc.expectErr && err != nil {
			t.Fatalf("%s: expected: %v, got: %v,", tc.name, tc.want, got)
		}
		if !slices.Equal(tc.want, got) {
			t.Fatalf("%s: expected: %v, got: %v,", tc.name, tc.want, got)
		}
	}

}
