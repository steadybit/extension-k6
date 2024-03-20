package extk6

import (
	"context"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
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
