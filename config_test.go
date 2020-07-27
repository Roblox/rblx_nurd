package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	err := loadConfig("NOPATH")
	assert.Error(t, err)
	assert.Empty(t, nomadAddresses)
	assert.Empty(t, metricsAddress)

	err = loadConfig("config_test.json")
	assert.Empty(t, err)
	assert.Equal(t, 2, len(nomadAddresses))
	assert.IsType(t, []string{}, nomadAddresses)
	assert.IsType(t, "", nomadAddresses[0])
	assert.Equal(t, "NomadURL0:NomadPort0", nomadAddresses[0])
	assert.IsType(t, "", nomadAddresses[1])
	assert.Equal(t, "NomadURL1:NomadPort1", nomadAddresses[1])
	assert.IsType(t, "", metricsAddress)
	assert.Equal(t, "VMURL:VMPort", metricsAddress)
}