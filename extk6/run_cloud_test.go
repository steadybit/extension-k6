package extk6

import (
	"context"
	"github.com/google/uuid"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/extension-k6/config"
	"github.com/steadybit/extension-kit/extutil"
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
	action := K6LoadTestCloudAction{}
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
	action := K6LoadTestCloudAction{}
	state := action.NewEmptyState()
	state.Command = []string([]string{"k6-not-available", "cloud", "test.js"})
	state.ExecutionId = uuid.New()
	config.Config.CloudApiToken = "test1234567890"
	// When
	result, err := action.Start(context.TODO(), &state)

	// Then
	require.Nil(t, result)
	require.Equal(t, err.Error(), "Failed to start command.")
}
