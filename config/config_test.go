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
