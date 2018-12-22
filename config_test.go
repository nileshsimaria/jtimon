package main

import (
	"testing"
)

func TestNewJTIMONConfig(t *testing.T) {
	var xerr error
	tests := []struct {
		name  string
		error bool
	}{
		{"tests/data/noerror.json", false}, // no error
		{"tests/data/error.jso", true},     // file does not exists
		{"tests/data/error.json", true},    // syntax error in JSON file
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewJTIMONConfig(test.name)
			if err != nil && !test.error {
				t.Errorf("NewJTIMONConfig failed, got: %v, want: %v", err, xerr)
			}
			if err == nil && test.error {
				t.Errorf("NewJTIMONConfig failed, got: %v, want error", err)
			}
		})
	}
}

func TestNewJTIMONConfigFilelist(t *testing.T) {
	var xerr error
	tests := []struct {
		name  string
		error bool
	}{
		{"tests/data/file_list.json", false},    // no error
		{"tests/data/file_list.jso", true},      // file does not exists
		{"tests/data/file_list_err.json", true}, // syntax error in JSON file
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewJTIMONConfigFilelist(test.name)
			if err != nil && !test.error {
				t.Errorf("NewJTIMONConfig failed, got: %v, want: %v", err, xerr)
			}
			if err == nil && test.error {
				t.Errorf("NewJTIMONConfig failed, got: %v, want error", err)
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name  string
		error bool
	}{
		{"tests/data/noerror.json", false}, // no error
		{"tests/data/error.jso", true},     // file does not exists
		{"tests/data/error.json", true},    // syntax error in JSON file
	}

	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			config, err := NewJTIMONConfig(test.name)
			if err != nil {
				configString, err := ValidateConfig(config)
				if err != nil {
					t.Errorf("TestValidateConfig failed. Error: %v\nConfig %s\n", err, configString)
				}
			}
		})
	}
}

func TestExploreConfig(t *testing.T) {
	t.Run("explore-config", func(t *testing.T) {
		_, err := ExploreConfig()
		if err != nil {
			t.Errorf("ExploreConfig failed, Error: %v\n", err)
		}
	})
}

func TestStringInSlice(t *testing.T) {
	tests := []struct {
		name  string
		value bool
	}{
		{"first", false},
		{"two", true},
		{"third", false},
	}

	configfilelist := []string{"one", "two", "three"}

	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			ret := StringInSlice(test.name, configfilelist)
			if ret != test.value {
				t.Errorf("TeststringInSlice failed")
			}
		})
	}
}
