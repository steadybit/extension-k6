// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package e2e

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"os"
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
		{
			Name: "run k6 archive",
			Test: testRunK6Archive,
		},
		{
			Name: "cancel k6",
			Test: testCancelK6,
		},
	})
}

func testRunK6(t *testing.T, _ *e2e.Minikube, e *e2e.Extension) {
	runK6(t, e, client.File{
		ParameterName: "file",
		FileName:      "script.js",
		Content:       []byte("import http from 'k6/http';\nexport default function() { http.get('https://www.steadybit.com'); }"),
	})
}

// testRunK6Archive runs a script that imports another local file, uploaded as the archive
// `k6 archive` produced from testdata/archive (see testdata/README.md to regenerate it).
func testRunK6Archive(t *testing.T, _ *e2e.Minikube, e *e2e.Extension) {
	content, err := os.ReadFile("testdata/archive.tar")
	require.NoError(t, err)
	metrics := runK6(t, e, client.File{
		ParameterName: "file",
		FileName:      "archive.tar",
		Content:       content,
	})
	// The imported module records this counter itself, so it is only there if the archive's
	// relative import resolved and ran.
	assert.GreaterOrEqual(t, metrics["archive_module_calls"], 1.0,
		"the module imported via ../lib/requests.js must have run")
}

// testCancelK6 checks that cancelling a running load test stops k6. The script is meant to run far
// longer than this test, so k6 is reliably running when the test looks for it: the long run is the
// scenario under test, not a timing allowance. Cancel only returns once the extension's stop has
// killed k6, so the process must be gone from then on.
func testCancelK6(t *testing.T, m *e2e.Minikube, e *e2e.Extension) {
	config := struct{}{}
	files := []client.File{{
		ParameterName: "file",
		FileName:      "script.js",
		Content: []byte("import http from 'k6/http';\n" +
			"import { sleep } from 'k6';\n" +
			"export const options = { vus: 1, duration: '10m' };\n" +
			"export default function() { http.get('https://www.steadybit.com'); sleep(1); }"),
	}}
	exec, err := e.RunActionWithFiles("com.steadybit.extension_k6.run", nil, config, nil, files)
	require.NoError(t, err)
	e2e.AssertProcessRunningInContainer(t, m, e.Pod, "extension", "k6", true)

	require.NoError(t, exec.Cancel())
	e2e.AssertProcessNOTRunningInContainer(t, m, e.Pod, "extension", "k6")
	assert.Contains(t, artifactsByLabel(t, exec), k6LogArtifact, "stopping must still return k6's log")
}

// runK6 runs the action with the given file to completion and checks what it returned: a clean
// completion, k6's log and the metrics k6 recorded. All of these exist once the action has ended,
// so nothing here depends on how long k6 runs. Returns the recorded metrics, summed per name.
func runK6(t *testing.T, e *e2e.Extension, file client.File) map[string]float64 {
	config := struct{}{}
	files := []client.File{file}
	exec, err := e.RunActionWithFiles("com.steadybit.extension_k6.run", nil, config, nil, files)
	require.NoError(t, err)
	require.NoError(t, exec.Wait(), "the action fails if k6 exits with an error")

	artifacts := artifactsByLabel(t, exec)
	require.Contains(t, artifacts, k6LogArtifact)
	require.Contains(t, artifacts, k6MetricsArtifact)
	assert.NotEmpty(t, artifacts[k6LogArtifact], "the log artifact must carry k6's output")

	metrics := sumK6Metrics(t, artifacts[k6MetricsArtifact])
	assert.GreaterOrEqual(t, metrics["iterations"], 1.0, "k6 must have completed an iteration of the script")
	assert.GreaterOrEqual(t, metrics["http_reqs"], 1.0, "the script's request must have been made")
	return metrics
}

func artifactsByLabel(t *testing.T, exec client.ActionExecution) map[string][]byte {
	t.Helper()
	artifacts := make(map[string][]byte)
	for _, artifact := range exec.Artifacts() {
		data, err := base64.StdEncoding.DecodeString(artifact.Data)
		require.NoError(t, err, "artifact %s must be valid base64", artifact.Label)
		artifacts[artifact.Label] = data
	}
	return artifacts
}

// sumK6Metrics parses the newline-delimited JSON k6 writes with --out json and sums the values of
// all data points per metric name.
func sumK6Metrics(t *testing.T, data []byte) map[string]float64 {
	t.Helper()
	sums := make(map[string]float64)
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		var line struct {
			Type   string `json:"type"`
			Metric string `json:"metric"`
			Data   struct {
				Value float64 `json:"value"`
			} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(scanner.Bytes(), &line), "the metrics artifact must be k6's JSON output")
		if line.Type == "Point" {
			sums[line.Metric] += line.Data.Value
		}
	}
	require.NoError(t, scanner.Err())
	return sums
}
