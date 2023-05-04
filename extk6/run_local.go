/*
 * Copyright 2023 steadybit GmbH. All rights reserved.
 */

package extk6

import (
	"context"
	"fmt"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/action-kit/go/action_kit_sdk"
	extension_kit "github.com/steadybit/extension-kit"
	"github.com/steadybit/extension-kit/extconversion"
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
	return *getActionDescription(fmt.Sprintf("%s.run", actionIdPrefix), "K6", "Execute a K6 load test.")
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
	return start(state)
}

func (l *K6LoadTestRunAction) Status(_ context.Context, state *K6LoadTestRunState) (*action_kit_api.StatusResult, error) {
	return status(state)
}

func (l *K6LoadTestRunAction) Stop(_ context.Context, state *K6LoadTestRunState) (*action_kit_api.StopResult, error) {
	return stop(state)
}
