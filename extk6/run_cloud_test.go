/*
 * Copyright 2024 steadybit GmbH. All rights reserved.
 */

package extk6

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/jarcoal/httpmock"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/extension-k6/config"
	"github.com/steadybit/extension-kit/extutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPrepareExtractsState(t *testing.T) {
	// Given
	request := extutil.JsonMangle(action_kit_api.PrepareActionRequestBody{
		Config: map[string]interface{}{
			"duration": 1000 * 60,
			"notify":   true,
			"file":     "test.js",
		},
		Target: nil,
		ExecutionContext: extutil.Ptr(action_kit_api.ExecutionContext{
			ExperimentUri: extutil.Ptr("<uri-to-experiment>"),
			ExecutionUri:  extutil.Ptr("<uri-to-execution>"),
		}),
	})
	action := k6LoadTestCloudAction{}
	state := action.NewEmptyState()

	// When
	result, err := action.Prepare(context.TODO(), &state, request)

	// Then
	require.Nil(t, result)
	require.Nil(t, err)
	require.Equal(t, state.Command, []string([]string{"k6", "cloud", "test.js"}))
}

func TestFailedCommandStart(t *testing.T) {
	// Given
	action := k6LoadTestCloudAction{}
	state := action.NewEmptyState()
	state.Command = []string([]string{"k6-not-available", "cloud", "test.js"})
	state.ExecutionId = uuid.New()
	config.Config.CloudApiToken = "test1234567890"
	// When
	result, err := action.Start(context.TODO(), &state)

	// Then
	require.Nil(t, result)
	require.Equal(t, "Failed to start command.: exec: \"k6-not-available\": executable file not found in $PATH", err.Error())
}

func Test_isCloudRunStillRunning(t *testing.T) {
	config.ParseConfiguration()
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", "https://api.k6.io/loadtests/v2/runs/2-running", httpmock.NewStringResponder(200, "{\"k6-run\": {\"run_status\": 2}}"))
	httpmock.RegisterResponder("GET", "https://api.k6.io/loadtests/v2/runs/3-finished", httpmock.NewStringResponder(200, "{\"k6-run\": {\"run_status\": 3}}"))
	httpmock.RegisterResponder("GET", "https://api.k6.io/loadtests/v2/runs/5-aborted", httpmock.NewStringResponder(200, "{\"k6-run\": {\"run_status\": 5}}"))
	httpmock.RegisterResponder("GET", "https://api.k6.io/loadtests/v2/runs/98-error", httpmock.NewErrorResponder(fmt.Errorf("error")))
	httpmock.RegisterResponder("GET", "https://api.k6.io/loadtests/v2/runs/99-invalid", httpmock.NewStringResponder(200, "invalid json"))
	httpmock.RegisterNoResponder(httpmock.NewStringResponder(404, ""))

	tests := []struct {
		name       string
		cloudRunId string
		want       bool
		wantErr    assert.ErrorAssertionFunc
	}{
		{
			name:       "Running",
			cloudRunId: "2-running",
			want:       true,
			wantErr:    assert.NoError,
		},
		{
			name:       "Finished",
			cloudRunId: "3-finished",
			want:       false,
			wantErr:    assert.NoError,
		},
		{
			name:       "Aborted",
			cloudRunId: "5-aborted",
			want:       false,
			wantErr:    assert.NoError,
		},
		{
			name:       "Not Found",
			cloudRunId: "404-not-found",
			want:       false,
			wantErr:    assert.Error,
		},
		{
			name:       "Error",
			cloudRunId: "98-error",
			want:       false,
			wantErr:    assert.Error,
		},
		{
			name:       "Invalid",
			cloudRunId: "99-invalid",
			want:       false,
			wantErr:    assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := isCloudRunStillRunning(tt.cloudRunId)
			if !tt.wantErr(t, err, fmt.Sprintf("isCloudRunStillRunning(= %v)", tt.cloudRunId)) {
				return
			}
			if got != tt.want {
				t.Errorf("isCloudRunStillRunning() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_stopCloudRun(t *testing.T) {
	config.ParseConfiguration()
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", "https://api.k6.io/loadtests/v2/runs/ok/stop", httpmock.NewStringResponder(200, "{\"k6-run\": {\"run_status\": 2}}"))
	httpmock.RegisterResponder("POST", "https://api.k6.io/loadtests/v2/runs/failure/stop", httpmock.NewStringResponder(500, ""))
	httpmock.RegisterResponder("POST", "https://api.k6.io/loadtests/v2/runs/error/stop", httpmock.NewErrorResponder(fmt.Errorf("error")))
	httpmock.RegisterNoResponder(httpmock.NewStringResponder(404, ""))

	tests := []struct {
		name       string
		cloudRunId string
		wantErr    assert.ErrorAssertionFunc
	}{
		{
			name:       "OK",
			cloudRunId: "ok",
			wantErr:    assert.NoError,
		},
		{
			name:       "Not Found",
			cloudRunId: "not-found",
			wantErr:    assert.NoError,
		},
		{
			name:       "ServerError",
			cloudRunId: "failure",
			wantErr:    assert.NoError,
		},
		{
			name:       "error",
			cloudRunId: "error",
			wantErr:    assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, stopCloudRun(tt.cloudRunId), fmt.Sprintf("stopCloudRun(%v)", tt.cloudRunId))
		})
	}
}
