package main

import (
	"reflect"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	err := loadConfig("NOPATH")
	if err == nil {
		t.Errorf("Value incorrect, got: %v, want %v", err, "open NOPATH: no such file or directory")
	}
	if len(nomadAddresses) != 0 {
		t.Errorf("Length incorrect, got: %v, want %v", len(nomadAddresses), 0)
	}
	if metricsAddress != "" {
		t.Errorf("Value incorrect, got: %v, want %v", metricsAddress, "")
	}

	err = loadConfig("config_test.json")
	if err != nil {
		t.Errorf("Value incorrect, got: %v, want %v", err, "<nil>")
	}
	if len(nomadAddresses) != 2 {
		t.Errorf("Length incorrect, got: %v, want %v", len(nomadAddresses), 2)
	}
	if reflect.TypeOf(nomadAddresses).Kind() != reflect.Slice {
		t.Errorf("Type incorrect, got: %T, want %v", nomadAddresses, "slice")
	}
	if reflect.TypeOf(nomadAddresses[0]).Kind() != reflect.String {
		t.Errorf("Type incorrect, got: %T, want %v", nomadAddresses[0], "string")
	}
	if nomadAddresses[0] != "NomadURL0:NomadPort0" {
		t.Errorf("Value incorrect, got: %v, want %v", nomadAddresses[0], "NomadURL0:NomadPort0")
	}
	if reflect.TypeOf(nomadAddresses[1]).Kind() != reflect.String {
		t.Errorf("Type incorrect, got: %T, want %v", nomadAddresses[1], "string")
	}
	if nomadAddresses[1] != "NomadURL1:NomadPort1" {
		t.Errorf("Value incorrect, got: %v, want %v", nomadAddresses[1], "NomadURL1:NomadPort1")
	}
	if reflect.TypeOf(metricsAddress).Kind() != reflect.String {
		t.Errorf("Type incorrect, got: %T, want %v", metricsAddress, "string")
	}
	if metricsAddress != "VMURL:VMPort" {
		t.Errorf("Value incorrect, got: %v, want %v", metricsAddress, "VMURL:VMPort")
	}
}