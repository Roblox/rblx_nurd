/*
Copyright 2020 Roblox Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

	
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
	assert.IsType(t, []string{}, nomadAddresses)
	assert.Equal(t, 2, len(nomadAddresses))
	assert.IsType(t, "", nomadAddresses[0])
	assert.Equal(t, "NomadURL0:NomadPort0", nomadAddresses[0])
	assert.IsType(t, "", nomadAddresses[1])
	assert.Equal(t, "NomadURL1:NomadPort1", nomadAddresses[1])
	assert.IsType(t, "", metricsAddress)
	assert.Equal(t, "VMURL:VMPort", metricsAddress)
}