/*
 * Copyright 2026 steadybit GmbH. All rights reserved.
 */

package extk6

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jarcoal/httpmock"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/action-kit/go/action_kit_sdk"
	"github.com/steadybit/extension-k6/config"
	"github.com/steadybit/extension-kit/extutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newExecutionDir creates the per-execution folder, which is managed by the action_kit_sdk in production.
func newExecutionDir(t *testing.T) uuid.UUID {
	t.Helper()
	executionId := uuid.New()
	dir := fmt.Sprintf("/tmp/steadybit/%v", executionId)
	require.NoError(t, os.MkdirAll(dir, 0755))
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	return executionId
}

// startFakeK6 starts a shell script in place of k6.
func startFakeK6(t *testing.T, script string) *K6LoadTestRunState {
	t.Helper()
	state := &K6LoadTestRunState{
		Command:     []string{"sh", "-c", script},
		ExecutionId: newExecutionDir(t),
	}
	result, err := start(state, "")
	require.NoError(t, err)
	require.Nil(t, result)
	require.NotEmpty(t, state.CmdStateID)
	require.Nil(t, state.Command, "command is cleared after start")
	return state
}

// awaitCompletion polls the status until the run is completed and returns the last result and all messages.
func awaitCompletion(t *testing.T, state *K6LoadTestRunState) (*action_kit_api.StatusResult, []string) {
	t.Helper()
	var messages []string
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		result, err := status(state)
		require.NoError(t, err)
		for _, m := range *result.Messages {
			messages = append(messages, m.Message)
		}
		if result.Completed {
			return result, messages
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("k6 run did not complete")
	return nil, nil
}

func Test_lifecycle_success(t *testing.T) {
	state := startFakeK6(t, `echo "running"; echo "output: https://app.k6.io/runs/4711"; echo ""; echo "done"`)

	result, messages := awaitCompletion(t, state)
	assert.Nil(t, result.Error)
	assert.Equal(t, []string{"running", "output: https://app.k6.io/runs/4711", "done"}, messages)
	assert.Equal(t, "4711", state.CloudRunId)

	log, err := os.ReadFile(fmt.Sprintf("/tmp/steadybit/%v/k6_log.txt", state.ExecutionId))
	require.NoError(t, err)
	assert.Contains(t, string(log), "running")

	require.NoError(t, os.WriteFile(fmt.Sprintf("/tmp/steadybit/%v/metrics.json", state.ExecutionId), []byte(`{"metric":"http_reqs"}`), 0644))
	stopResult, err := stop(state)
	require.NoError(t, err)
	require.Len(t, *stopResult.Artifacts, 2)
	assert.Equal(t, "$(experimentKey)_$(executionId)_k6_log.txt", (*stopResult.Artifacts)[0].Label)
	assert.Equal(t, "$(experimentKey)_$(executionId)_k6_metrics.json", (*stopResult.Artifacts)[1].Label)
	for _, m := range *stopResult.Messages {
		assert.NotEqual(t, action_kit_api.Error, *m.Level, "unexpected error message %q", m.Message)
	}
}

func Test_lifecycle_threshold_failure(t *testing.T) {
	for _, exitCode := range []int{97, 99} {
		t.Run(fmt.Sprintf("exit code %d", exitCode), func(t *testing.T) {
			state := startFakeK6(t, fmt.Sprintf("echo 'thresholds crossed'; exit %d", exitCode))

			result, _ := awaitCompletion(t, state)
			require.NotNil(t, result.Error)
			assert.Equal(t, "Some thresholds have failed.", result.Error.Title)
			assert.Equal(t, extutil.Ptr(action_kit_api.Failed), result.Error.Status)

			stopResult, err := stop(state)
			require.NoError(t, err)
			messages := *stopResult.Messages
			require.NotEmpty(t, messages)
			assert.Equal(t, action_kit_api.Error, *messages[len(messages)-1].Level)
			assert.Equal(t, fmt.Sprintf("K6 run failed with exit code %d", exitCode), messages[len(messages)-1].Message)
		})
	}
}

func Test_lifecycle_error(t *testing.T) {
	t.Run("error from log", func(t *testing.T) {
		state := startFakeK6(t, `echo 'time="2026" level=error msg=script not found'; exit 107`)

		result, _ := awaitCompletion(t, state)
		require.NotNil(t, result.Error)
		// the title currently keeps the line break of the log line
		assert.Equal(t, "script not found", strings.TrimSpace(result.Error.Title))
		assert.Equal(t, extutil.Ptr(action_kit_api.Errored), result.Error.Status)
	})

	t.Run("exit code only", func(t *testing.T) {
		state := startFakeK6(t, `exit 3`)

		result, _ := awaitCompletion(t, state)
		require.NotNil(t, result.Error)
		assert.Equal(t, "K6 run failed, exit-code 3", result.Error.Title)
	})
}

func Test_lifecycle_stop_while_running(t *testing.T) {
	state := startFakeK6(t, `echo "started"; exec sleep 30`)

	result, err := status(state)
	require.NoError(t, err)
	assert.False(t, result.Completed)

	stopResult, err := stop(state)
	require.NoError(t, err)
	require.NotNil(t, stopResult)
	assert.Len(t, *stopResult.Artifacts, 1, "only the log file exists, there are no metrics")

	_, err = status(state)
	assert.Error(t, err, "command state is removed after stop")
}

func Test_stop_not_started(t *testing.T) {
	result, err := stop(&K6LoadTestRunState{})
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func Test_unknown_command_state(t *testing.T) {
	state := &K6LoadTestRunState{CmdStateID: "unknown"}

	_, err := status(state)
	assert.ErrorContains(t, err, "Failed to find command state")

	_, err = stop(state)
	assert.ErrorContains(t, err, "Failed to find command state")
}

func Test_status_log_file_not_writable(t *testing.T) {
	state := &K6LoadTestRunState{
		Command: []string{"sh", "-c", "exit 0"},
		// no execution folder is created
		ExecutionId: uuid.New(),
	}
	_, err := start(state, "token")
	require.NoError(t, err)
	t.Cleanup(func() { _, _ = stop(state) })

	_, err = status(state)
	assert.ErrorContains(t, err, "Failed to append log to file")
}

func Test_start_passes_cloud_token(t *testing.T) {
	state := &K6LoadTestRunState{
		Command:     []string{"sh", "-c", `echo "token=$K6_CLOUD_TOKEN"`},
		ExecutionId: newExecutionDir(t),
	}
	_, err := start(state, "secret-token")
	require.NoError(t, err)

	_, messages := awaitCompletion(t, state)
	assert.Equal(t, []string{"token=secret-token"}, messages)
}

func Test_stdOutToMessages(t *testing.T) {
	messages := stdOutToMessages([]string{"  first\n", "", "\n", "second"})
	require.Len(t, messages, 2)
	assert.Equal(t, "first", messages[0].Message)
	assert.Equal(t, "second", messages[1].Message)
	assert.Equal(t, action_kit_api.Info, *messages[0].Level)
}

func Test_addCloudRunIdToState(t *testing.T) {
	state := &K6LoadTestRunState{}
	addCloudRunIdToState([]string{"execution: cloud", "  output: https://app.k6.io/runs/12345\n"}, state)
	assert.Equal(t, "12345", state.CloudRunId)

	state = &K6LoadTestRunState{}
	addCloudRunIdToState([]string{"execution: local", "output: https://app.k6.io/runs/"}, state)
	assert.Empty(t, state.CloudRunId)
}

func Test_prepare_invalid_config(t *testing.T) {
	state := &K6LoadTestRunState{}
	_, err := prepare(state, action_kit_api.PrepareActionRequestBody{
		Config: map[string]any{"environment": "not a list"},
	}, []string{"k6"})
	assert.ErrorContains(t, err, "Failed to unmarshal the config.")
}

func Test_localAction_Prepare(t *testing.T) {
	action := NewK6LoadTestRunAction()
	state := action.NewEmptyState()
	executionId := uuid.New()

	result, err := action.Prepare(context.Background(), &state, extutil.JsonMangle(action_kit_api.PrepareActionRequestBody{
		ExecutionId: executionId,
		Config: map[string]any{
			"file": "/tmp/script.js",
			"environment": []map[string]string{
				{"key": "HOST", "value": "shop.example.com"},
				{"key": "VUS", "value": "10"},
			},
		},
	}))
	require.NoError(t, err)
	assert.Nil(t, result)
	assert.Equal(t, executionId, state.ExecutionId)
	assert.Equal(t, []string{
		"k6", "run", "/tmp/script.js", "--no-usage-report", "--out", fmt.Sprintf("json=/tmp/steadybit/%v/metrics.json", executionId),
		"--env", "HOST=shop.example.com",
		"--env", "VUS=10",
	}, state.Command)
}

// A k6 archive is self-describing, so it is passed to k6 exactly like a single script.
func Test_Prepare_archive(t *testing.T) {
	tests := []struct {
		name    string
		action  action_kit_sdk.Action[K6LoadTestRunState]
		command []string
	}{
		{"local", NewK6LoadTestRunAction(), []string{"k6", "run", "/tmp/archive.tar"}},
		{"cloud", NewK6LoadTestCloudAction(), []string{"k6", "cloud", "run", "/tmp/archive.tar"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := tt.action.NewEmptyState()
			_, err := tt.action.Prepare(context.Background(), &state, extutil.JsonMangle(action_kit_api.PrepareActionRequestBody{
				ExecutionId: uuid.New(),
				Config:      map[string]any{"file": "/tmp/archive.tar"},
			}))
			require.NoError(t, err)
			assert.Equal(t, tt.command, state.Command[:len(tt.command)])
		})
	}
}

func Test_fileParameter_accepts_script_and_archive(t *testing.T) {
	for _, description := range []action_kit_api.ActionDescription{
		NewK6LoadTestRunAction().Describe(),
		NewK6LoadTestCloudAction().Describe(),
	} {
		file := description.Parameters[0]
		assert.Equal(t, "file", file.Name)
		assert.Equal(t, []string{".js", ".tar"}, *file.AcceptedFileTypes, description.Id)
	}
}

func Test_localAction_Lifecycle(t *testing.T) {
	action := &K6LoadTestRunAction{}
	state := K6LoadTestRunState{
		Command:     []string{"sh", "-c", "echo local"},
		ExecutionId: newExecutionDir(t),
	}

	_, err := action.Start(context.Background(), &state)
	require.NoError(t, err)

	var statusResult *action_kit_api.StatusResult
	require.Eventually(t, func() bool {
		statusResult, err = action.Status(context.Background(), &state)
		require.NoError(t, err)
		return statusResult.Completed
	}, 10*time.Second, 50*time.Millisecond)

	_, err = action.Stop(context.Background(), &state)
	require.NoError(t, err)
}

func Test_localAction_Describe(t *testing.T) {
	previous := config.Config
	t.Cleanup(func() { config.Config = previous })

	config.Config.EnableLocationSelection = false
	description := NewK6LoadTestRunAction().Describe()
	assert.Equal(t, "com.steadybit.extension_k6.run", description.Id)
	assert.Equal(t, action_kit_api.LoadTest, description.Kind)
	assert.NotNil(t, description.Hint)
	assert.Nil(t, description.TargetSelection)
	assert.Len(t, description.Parameters, 2)

	config.Config.EnableLocationSelection = true
	description = NewK6LoadTestRunAction().Describe()
	require.NotNil(t, description.TargetSelection)
	assert.Equal(t, targetType, description.TargetSelection.TargetType)
	assert.Len(t, description.Parameters, 3)
}

func Test_cloudAction_Describe(t *testing.T) {
	description := NewK6LoadTestCloudAction().Describe()
	assert.Equal(t, "com.steadybit.extension_k6.cloud", description.Id)
	assert.Nil(t, description.Hint)
	assert.Nil(t, description.TargetSelection)
}

func Test_cloudAction_Stop(t *testing.T) {
	config.ParseConfiguration()
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", "https://api.k6.io/loadtests/v2/runs/running", httpmock.NewStringResponder(200, `{"k6-run": {"run_status": 2}}`))
	httpmock.RegisterResponder("GET", "https://api.k6.io/loadtests/v2/runs/finished", httpmock.NewStringResponder(200, `{"k6-run": {"run_status": 3}}`))
	httpmock.RegisterResponder("GET", "https://api.k6.io/loadtests/v2/runs/broken", httpmock.NewStringResponder(500, ""))
	httpmock.RegisterResponder("POST", "https://api.k6.io/loadtests/v2/runs/running/stop", httpmock.NewStringResponder(200, "{}"))
	httpmock.RegisterResponder("POST", "https://api.k6.io/loadtests/v2/runs/stop-fails/stop", httpmock.NewErrorResponder(fmt.Errorf("connection refused")))
	httpmock.RegisterResponder("GET", "https://api.k6.io/loadtests/v2/runs/stop-fails", httpmock.NewStringResponder(200, `{"k6-run": {"run_status": 1}}`))

	action := &k6LoadTestCloudAction{}

	t.Run("stops running cloud run", func(t *testing.T) {
		httpmock.ZeroCallCounters()
		_, err := action.Stop(context.Background(), &K6LoadTestRunState{CloudRunId: "running"})
		require.NoError(t, err)
		assert.Equal(t, 1, httpmock.GetCallCountInfo()["POST https://api.k6.io/loadtests/v2/runs/running/stop"])
	})

	t.Run("does not stop finished cloud run", func(t *testing.T) {
		httpmock.ZeroCallCounters()
		_, err := action.Stop(context.Background(), &K6LoadTestRunState{CloudRunId: "finished"})
		require.NoError(t, err)
		assert.Equal(t, 0, httpmock.GetCallCountInfo()["POST https://api.k6.io/loadtests/v2/runs/finished/stop"])
	})

	t.Run("status request fails", func(t *testing.T) {
		_, err := action.Stop(context.Background(), &K6LoadTestRunState{CloudRunId: "broken"})
		assert.ErrorContains(t, err, "failed to get k6 cloud status")
	})

	t.Run("stop request fails", func(t *testing.T) {
		_, err := action.Stop(context.Background(), &K6LoadTestRunState{CloudRunId: "stop-fails"})
		assert.ErrorContains(t, err, "failed to stop k6 cloud")
	})

	t.Run("no cloud run id", func(t *testing.T) {
		result, err := action.Stop(context.Background(), &K6LoadTestRunState{})
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func Test_cloudAction_Status(t *testing.T) {
	action := &k6LoadTestCloudAction{}
	state := K6LoadTestRunState{
		Command:     []string{"sh", "-c", "echo cloud"},
		ExecutionId: newExecutionDir(t),
	}
	_, err := start(&state, "")
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		result, err := action.Status(context.Background(), &state)
		require.NoError(t, err)
		return result.Completed
	}, 10*time.Second, 50*time.Millisecond)
}

func Test_appendFileArtifact_zip_fails(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "k6_log.txt")
	require.NoError(t, os.WriteFile(path, make([]byte, artifactZipThreshold+1), 0644))
	// a directory with the name of the archive prevents creating it
	require.NoError(t, os.Mkdir(filepath.Join(dir, "k6_log.zip"), 0755))

	_, err := appendFileArtifact(nil, path, "k6_log.txt")
	assert.ErrorContains(t, err, "Failed to zip")
}
