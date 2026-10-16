/*
 * Copyright 2026 steadybit GmbH. All rights reserved.
 */

package extk6

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/action-kit/go/action_kit_sdk"
	"github.com/steadybit/extension-kit/extutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubInspect replaces `k6 inspect` with a fixed result and returns the arguments of each call.
func stubInspect(t *testing.T, output string, err error) *[][]string {
	t.Helper()
	previous := inspectK6
	t.Cleanup(func() { inspectK6 = previous })
	calls := &[][]string{}
	inspectK6 = func(_ context.Context, args []string) ([]byte, error) {
		*calls = append(*calls, args)
		return []byte(output), err
	}
	return calls
}

func prepareWith(t *testing.T, action action_kit_sdk.Action[K6LoadTestRunState], config map[string]any) (K6LoadTestRunState, *action_kit_api.PrepareResult) {
	t.Helper()
	state := action.NewEmptyState()
	result, err := action.Prepare(context.Background(), &state, extutil.JsonMangle(action_kit_api.PrepareActionRequestBody{
		ExecutionId: uuid.New(),
		Config:      config,
	}))
	require.NoError(t, err)
	return state, result
}

func Test_Prepare_scenarios_conflict(t *testing.T) {
	for _, config := range []map[string]any{
		{"file": "/tmp/script.js", "duration": 30000},
		{"file": "/tmp/script.js", "K6_VUS": 10},
	} {
		stubInspect(t, `{"scenarios": {"a": {"executor": "constant-vus"}}}`, nil)
		state, result := prepareWith(t, NewK6LoadTestRunAction(), config)

		require.NotNil(t, result)
		require.NotNil(t, result.Error)
		assert.Equal(t, "The VUs and Duration parameters can't be combined with k6 scenarios. Clear the parameters or remove the scenarios.", result.Error.Title)
		assert.Equal(t, extutil.Ptr(action_kit_api.Errored), result.Error.Status)
		assert.Nil(t, state.Command, "nothing is run")
	}
}

func Test_Prepare_scenarios_not_checked_without_overrides(t *testing.T) {
	calls := stubInspect(t, `{"scenarios": {"a": {}}}`, nil)
	state, result := prepareWith(t, NewK6LoadTestRunAction(), map[string]any{"file": "/tmp/script.js"})

	assert.Nil(t, result)
	assert.NotEmpty(t, state.Command)
	assert.Empty(t, *calls, "a script with scenarios runs as before when duration and VUs are not set")
}

func Test_Prepare_scenarios_inspect_arguments(t *testing.T) {
	config := map[string]any{
		"file":        "/tmp/script.js",
		"duration":    30000,
		"environment": []map[string]string{{"key": "PROFILE", "value": "smoke"}},
	}
	for _, action := range []action_kit_sdk.Action[K6LoadTestRunState]{NewK6LoadTestRunAction(), NewK6LoadTestCloudAction()} {
		calls := stubInspect(t, `{"scenarios": null}`, nil)
		prepareWith(t, action, config)
		assert.Equal(t, [][]string{{"--env", "PROFILE=smoke", "/tmp/script.js"}}, *calls, action.Describe().Id)
	}
}

func Test_Prepare_scenarios_check_skipped_when_inspect_fails(t *testing.T) {
	tests := []struct {
		name   string
		output string
		err    error
	}{
		{"inspect fails", "", errors.New("exit status 107")},
		{"unexpected output", "not json", nil},
		{"empty scenarios", `{"scenarios": {}}`, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stubInspect(t, tt.output, tt.err)
			state, result := prepareWith(t, NewK6LoadTestRunAction(), map[string]any{"file": "/tmp/script.js", "duration": 30000})

			assert.Nil(t, result)
			assert.Contains(t, state.Command, "K6_DURATION=30s", "the run goes ahead and reports any script problem itself")
		})
	}
}

// Runs the real k6 binary, so the arguments and the JSON shape of `k6 inspect` are covered too.
func Test_scriptDefinesScenarios_with_k6(t *testing.T) {
	if _, err := exec.LookPath("k6"); err != nil {
		t.Skip("k6 is not installed")
	}
	dir := t.TempDir()
	write := func(name, content string) string {
		path := filepath.Join(dir, name)
		require.NoError(t, os.WriteFile(path, []byte(content), 0644))
		return path
	}
	plain := write("plain.js", "export const options = { vus: 2, duration: '5s' };\nexport default function () {}\n")
	scenarios := write("scenarios.js", "export const options = { scenarios: { a: { executor: 'constant-vus', vus: 1, duration: '3s' } } };\nexport default function () {}\n")
	byEnv := write("by-env.js", "export const options = __ENV.WITH_SCENARIOS ? { scenarios: { a: { executor: 'constant-vus', vus: 1, duration: '3s' } } } : {};\nexport default function () {}\n")

	assert.False(t, scriptDefinesScenarios(plain, nil))
	assert.True(t, scriptDefinesScenarios(scenarios, nil))
	assert.False(t, scriptDefinesScenarios(byEnv, nil))
	assert.True(t, scriptDefinesScenarios(byEnv, []string{"--env", "WITH_SCENARIOS=1"}))
	assert.False(t, scriptDefinesScenarios(filepath.Join(dir, "missing.js"), nil))

	archive := filepath.Join(dir, "scenarios.tar")
	require.NoError(t, exec.Command("k6", "archive", scenarios, "-O", archive).Run())
	assert.True(t, scriptDefinesScenarios(archive, nil), "the options stored in an archive are inspected too")
}
