package main

import (
	"testing"
)

func TestNewJTIMONConfig(t *testing.T) {
	var xerr error
	files := []struct {
		name  string
		error bool
	}{
		{"tests/data/noerror.json", false}, // no error
		{"tests/data/error.jso", true},     // file does not exists
		{"tests/data/error.json", true},    // syntax error in JSON file
	}

	for _, file := range files {
		_, err := NewJTIMONConfig(file.name)
		if err != nil && !file.error {
			t.Errorf("NewJTIMONConfig failed, got: %v, want: %v", err, xerr)
		}
		if err == nil && file.error {
			t.Errorf("NewJTIMONConfig failed, got: %v, want error", err)
		}
	}
}

func TestValidateConfig(t *testing.T) {
	files := []struct {
		name  string
		error bool
	}{
		{"tests/data/noerror.json", false}, // no error
		{"tests/data/error.jso", true},     // file does not exists
		{"tests/data/error.json", true},    // syntax error in JSON file
	}

	for _, file := range files {
		config, err := NewJTIMONConfig(file.name)
		if err != nil {
			configString, err := ValidateConfig(config)
			if err != nil {
				t.Errorf("TestValidateConfig failed. Error: %v\nConfig %s\n", err, configString)
			}
		}
	}
}

func TestExploreConfig(t *testing.T) {
	_, err := ExploreConfig()
	if err != nil {
		t.Errorf("ExploreConfig failed, Error: %v\n", err)
	}
}
