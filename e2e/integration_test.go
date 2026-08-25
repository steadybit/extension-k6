// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package e2e

import (
	"bytes"
	"encoding/base64"
	"testing"

	"github.com/steadybit/action-kit/go/action_kit_test/client"
	"github.com/steadybit/action-kit/go/action_kit_test/e2e"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The $(...) placeholders are substituted by the platform, so they arrive verbatim here.
const (
	k6LogArtifact     = "$(experimentKey)_$(executionId)_k6_log.txt"
	k6MetricsArtifact = "$(experimentKey)_$(executionId)_k6_metrics.json"
)

func TestWithMinikube(t *testing.T) {
	extFactory := e2e.HelmExtensionFactory{
		Name: "extension-k6",
		Port: 8087,
		ExtraArgs: func(m *e2e.Minikube) []string {
			return []string{"--set", "logging.level=debug"}
		},
	}

	e2e.WithDefaultMinikube(t, &extFactory, []e2e.WithMinikubeTestCase{
		{
			Name: "run k6",
			Test: testRunK6,
		},
	})
}

func testRunK6(t *testing.T, m *e2e.Minikube, e *e2e.Extension) {
	config := struct{}{}
	files := []client.File{
		{
			ParameterName: "file",
			FileName:      "script.js",
			Content:       []byte("import http from 'k6/http';\nexport default function() { http.get('https://www.steadybit.com'); }"),
		},
	}
	exec, err := e.RunActionWithFiles("com.steadybit.extension_k6.run", nil, config, nil, files)
	require.NoError(t, err)
	e2e.AssertProcessRunningInContainer(t, m, e.Pod, "extension", "k6", true)
	///        \      Grafana   /‾‾/                                                                                                                                                                                                                                                                                                                                                                                                                  │
	//   /\  /  \     |\  __   /  /                                                                                                                                                                                                                                                                                                                                                                                                              │
	//  /  \/    \    | |/ /  /   ‾‾\                                                                                                                                                                                                                                                                                                                                                                                                           │
	// /          \   |   (  |  (‾)  |                                                                                                                                                                                                                                                                                                                                                                                                         │
	/// __________ \  |_|\_\  \_____/
	e2e.AssertLogContains(t, m, e.Pod, "/ __________ \\  |_|\\_\\  \\_____/")

	err = exec.Wait()
	require.NoError(t, err)

	// The banner asserted above only shows that k6 started, not that its output came back.
	artifacts := make(map[string][]byte)
	for _, artifact := range exec.Artifacts() {
		data, err := base64.StdEncoding.DecodeString(artifact.Data)
		require.NoError(t, err, "artifact %s must be valid base64", artifact.Label)
		artifacts[artifact.Label] = data
	}

	require.Contains(t, artifacts, k6LogArtifact)
	require.Contains(t, artifacts, k6MetricsArtifact)
	assert.NotEmpty(t, artifacts[k6LogArtifact], "the log artifact must carry k6's output")
	// k6 was pointed at this file with --out json=, so it holds newline-delimited JSON objects.
	assert.True(t, bytes.HasPrefix(artifacts[k6MetricsArtifact], []byte("{")),
		"the metrics artifact must be the json stream k6 was told to write")
}
