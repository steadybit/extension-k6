// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 Steadybit GmbH

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseConfiguration(t *testing.T) {
	previous := Config
	t.Cleanup(func() { Config = previous })

	t.Setenv("STEADYBIT_EXTENSION_ENABLE_LOCATION_SELECTION", "true")
	t.Setenv("STEADYBIT_EXTENSION_CLOUD_API_TOKEN", "token")
	t.Setenv("STEADYBIT_EXTENSION_KUBERNETES_NAMESPACE", "steadybit")

	ParseConfiguration()
	ValidateConfiguration()

	assert.True(t, Config.EnableLocationSelection)
	assert.Equal(t, "token", Config.CloudApiToken)
	assert.Equal(t, "steadybit", Config.KubernetesNamespace)
	assert.Equal(t, "https://api.k6.io", Config.CloudApiBaseUrl)
}

func TestParseConfiguration_k6EnvAllowList(t *testing.T) {
	previous := Config
	t.Cleanup(func() { Config = previous })

	t.Setenv("STEADYBIT_EXTENSION_K6_ENV_ALLOW_LIST", "MY_APP_URL, OTHER_VAR,,STEADYBIT_EXTENSION_CLOUD_API_TOKEN,steadybit_extension_foo,MY_APP_*")

	ParseConfiguration()
	ValidateConfiguration()

	assert.Equal(t, []string{"MY_APP_URL", "OTHER_VAR"}, Config.K6EnvAllowList)
}
