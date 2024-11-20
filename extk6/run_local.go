/*
 * Copyright 2023 steadybit GmbH. All rights reserved.
 */

package extk6

import (
	"context"
	"fmt"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/action-kit/go/action_kit_sdk"
	"github.com/steadybit/extension-k6/config"
	extension_kit "github.com/steadybit/extension-kit"
	"github.com/steadybit/extension-kit/extconversion"
	"github.com/steadybit/extension-kit/extutil"
)

type K6LoadTestRunAction struct{}

// Make sure action implements all required interfaces
var (
	_ action_kit_sdk.Action[K6LoadTestRunState]           = (*K6LoadTestRunAction)(nil)
	_ action_kit_sdk.ActionWithStatus[K6LoadTestRunState] = (*K6LoadTestRunAction)(nil)
	_ action_kit_sdk.ActionWithStop[K6LoadTestRunState]   = (*K6LoadTestRunAction)(nil)
)

func NewK6LoadTestRunAction() action_kit_sdk.Action[K6LoadTestRunState] {
	return &K6LoadTestRunAction{}
}

func (l *K6LoadTestRunAction) NewEmptyState() K6LoadTestRunState {
	return K6LoadTestRunState{}
}

func (l *K6LoadTestRunAction) Describe() action_kit_api.ActionDescription {
	hint := action_kit_api.ActionHint{
		Content: "Please note that load tests are executed by the k6 extension participating in the experiment, consuming resources of the system that it is installed in.",
		Type:    action_kit_api.HintWarning,
	}
	description := *getActionDescription(fmt.Sprintf("%s.run", actionIdPrefix), "K6", "Execute a K6 load test.", &hint)

	if config.Config.EnableLocationSelection {
		description.Parameters = append(description.Parameters, action_kit_api.ActionParameter{
			Name:  "-",
			Label: "Filter K6 Locations",
			Type:  action_kit_api.ActionParameterTypeTargetSelection,
			Order: extutil.Ptr(3),
		})
		description.TargetSelection = extutil.Ptr(action_kit_api.TargetSelection{
			TargetType: targetType,
			DefaultBlastRadius: extutil.Ptr(action_kit_api.DefaultBlastRadius{
				Mode:  action_kit_api.DefaultBlastRadiusModeMaximum,
				Value: 1,
			}),
			MissingQuerySelection: extutil.Ptr(action_kit_api.MissingQuerySelectionIncludeAll),
		})
	}
	return description
}

func (l *K6LoadTestRunAction) Prepare(_ context.Context, state *K6LoadTestRunState, request action_kit_api.PrepareActionRequestBody) (*action_kit_api.PrepareResult, error) {
	var config K6LoadTestRunConfig
	if err := extconversion.Convert(request.Config, &config); err != nil {
		return nil, extension_kit.ToError("Failed to unmarshal the config.", err)
	}
	filename := fmt.Sprintf("/tmp/steadybit/%v/metrics.json", request.ExecutionId) //Folder is managed by action_kit_sdk's file download handling
	command := []string{
		"k6",
		"run",
		config.File,
		"--no-usage-report",
		"--out",
		fmt.Sprintf("json=%s", filename),
	}
	return prepare(state, request, command)
}

func (l *K6LoadTestRunAction) Start(_ context.Context, state *K6LoadTestRunState) (*action_kit_api.StartResult, error) {
	return start(state, "")
}

func (l *K6LoadTestRunAction) Status(_ context.Context, state *K6LoadTestRunState) (*action_kit_api.StatusResult, error) {
	return status(state)
}

func (l *K6LoadTestRunAction) Stop(_ context.Context, state *K6LoadTestRunState) (*action_kit_api.StopResult, error) {
	return stop(state)
}
