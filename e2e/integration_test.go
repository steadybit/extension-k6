// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package e2e

import (
	"bytes"
	"encoding/base64"
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
	})
}

func testRunK6(t *testing.T, m *e2e.Minikube, e *e2e.Extension) {
	runK6(t, m, e, client.File{
		ParameterName: "file",
		FileName:      "script.js",
		Content: []byte("import http from 'k6/http';\n" +
			"import { sleep } from 'k6';\n" +
			"export const options = { vus: 1, duration: '3s' };\n" +
			"export default function() { http.get('https://www.steadybit.com'); sleep(1); }"),
	})
}

// testRunK6Archive runs a script that imports another local file, uploaded as the archive
// `k6 archive` produced from testdata/archive (see testdata/README.md to regenerate it).
// The archived script runs for 3s like the inline one, for the reason explained on runK6.
func testRunK6Archive(t *testing.T, m *e2e.Minikube, e *e2e.Extension) {
	content, err := os.ReadFile("testdata/archive.tar")
	require.NoError(t, err)
	runK6(t, m, e, client.File{
		ParameterName: "file",
		FileName:      "archive.tar",
		Content:       content,
	})
}

// runK6 runs the action with the given file and checks that k6 ran.
//
// The test scripts run for a fixed 3s (`vus: 1, duration: '3s'`, one request per second) so that
// AssertProcessRunningInContainer reliably finds the k6 process:
//   - RunActionWithFiles returns as soon as the action has started, so k6 is already running when
//     the assertion begins. The assertion then runs `ps` in the pod every 200ms (for up to 5s), and
//     each of those is a `kubectl exec` that itself takes a few hundred milliseconds.
//   - Without a duration, k6 runs a single iteration and exits after about 0.2s. It is then often
//     gone before the first `ps` lands, so whether the assertion passes depends on timing. That made
//     the test flaky, mostly for the archive case, which runs second on an already warm cluster.
//   - 3s keeps k6 running through the first polls with room for a slow `kubectl exec` on a busy CI
//     runner, while adding only a few seconds per test case.
func runK6(t *testing.T, m *e2e.Minikube, e *e2e.Extension, file client.File) {
	config := struct{}{}
	files := []client.File{file}
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
