/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"reflect"
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

func TestFilterResources(t *testing.T) {

	server := Resource{Id: "example_server", Data: Server{ID: "123456781234", Name: "example_server", Tags: []string{"hoge"}}}
	disk := Resource{Id: "example_disk", Data: Disk{ID: "123456788765", Name: "example_disk"}}
	resources := []Resource{
		server, disk,
	}
	tests := []struct {
		name         string
		searchFilter string
		resources    []Resource
		want         []Resource
	}{
		{name: "no filter", searchFilter: "", resources: resources, want: resources},
		{name: "1 searchWord", searchFilter: "hoge", resources: resources, want: []Resource{server}},
		{name: "AND condition", searchFilter: "example&disk", resources: resources, want: []Resource{disk}},
		{name: " OR condition", searchFilter: "server,disk", resources: resources, want: resources},
	}

	for _, tc := range tests {
		got, err := filterResource(tc.searchFilter, &tc.resources)

		if err != nil {
			t.Fatalf("%s: expected: %v, got: %v,", tc.name, tc.want, got)
		}
		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("%s: expected: %v, got: %v,", tc.name, tc.want, got)
		}
	}
}
