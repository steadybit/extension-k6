// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package e2e

import (
	"github.com/steadybit/action-kit/go/action_kit_test/client"
	"github.com/steadybit/action-kit/go/action_kit_test/e2e"
	"github.com/stretchr/testify/require"
	"testing"
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
	//          /\      |‾‾| /‾‾/   /‾‾/
	//     /\  /  \     |  |/  /   /  /
	//    /  \/    \    |     (   /   ‾‾\
	//   /          \   |  |\  \ |  (‾)  |
	//  / __________ \  |__| \__\ \_____/ .io
	e2e.AssertLogContains(t, m, e.Pod, "/ __________ \\  |__| \\__\\ \\_____/ .io")

	err = exec.Wait()
	require.NoError(t, err)
}
