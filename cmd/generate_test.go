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
		want            []string
		expectErr       bool
	}{
		{name: "include 1 service", services: []string{"server"}, excludeServices: []string{}, want: []string{"Server"}, expectErr: false},
		{name: "include multiple services", services: []string{"server", "disk"}, excludeServices: []string{}, want: []string{"Server", "Disk"}, expectErr: false},
		{name: "exclude 1 service", services: []string{}, excludeServices: []string{"server"}, want: []string{"Disk", "Switch"}, expectErr: false},
		{name: "exclude multiple services", services: []string{}, excludeServices: []string{"server", "disk"}, want: []string{"Switch"}, expectErr: false},
		{name: "specify no service", services: []string{}, excludeServices: []string{}, want: []string{"Server", "Disk", "Switch"}, expectErr: false},
		{name: "specify service and exclude service", services: []string{"server"}, excludeServices: []string{"disk"}, want: []string{}, expectErr: true},
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
