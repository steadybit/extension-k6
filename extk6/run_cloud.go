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
	"os/exec"
)

type K6LoadTestCloudAction struct{}

// Make sure action implements all required interfaces
var (
	_ action_kit_sdk.Action[K6LoadTestRunState]           = (*K6LoadTestCloudAction)(nil)
	_ action_kit_sdk.ActionWithStatus[K6LoadTestRunState] = (*K6LoadTestCloudAction)(nil)
	_ action_kit_sdk.ActionWithStop[K6LoadTestRunState]   = (*K6LoadTestCloudAction)(nil)
)

func NewK6LoadTestCloudAction() action_kit_sdk.Action[K6LoadTestRunState] {
	return &K6LoadTestCloudAction{}
}

func (l *K6LoadTestCloudAction) NewEmptyState() K6LoadTestRunState {
	return K6LoadTestRunState{}
}

func (l *K6LoadTestCloudAction) Describe() action_kit_api.ActionDescription {
	return *getActionDescription(fmt.Sprintf("%s.cloud", actionIdPrefix), "K6 Cloud", "Execute a K6 load using K6 Cloud.")
}

func (l *K6LoadTestCloudAction) Prepare(_ context.Context, state *K6LoadTestRunState, request action_kit_api.PrepareActionRequestBody) (*action_kit_api.PrepareResult, error) {
	var config K6LoadTestRunConfig
	if err := extconversion.Convert(request.Config, &config); err != nil {
		return nil, extension_kit.ToError("Failed to unmarshal the config.", err)
	}
	command := []string{
		"k6",
		"cloud",
		config.File,
	}
	return prepare(state, request, command)
}

func (l *K6LoadTestCloudAction) Start(_ context.Context, state *K6LoadTestRunState) (*action_kit_api.StartResult, error) {
	cmdLogin := exec.Command("k6", "login", "cloud", "-t", config.Config.CloudApiToken)
	if err := cmdLogin.Start(); err != nil {
		return nil, extension_kit.ToError("Failed to login to k6 cloud.", err)
	}
	if err := cmdLogin.Wait(); err != nil {
		return nil, extension_kit.ToError("Failed to login to k6 cloud.", err)
	}

	return start(state)
}

func (l *K6LoadTestCloudAction) Status(_ context.Context, state *K6LoadTestRunState) (*action_kit_api.StatusResult, error) {
	return status(state)
}

func (l *K6LoadTestCloudAction) Stop(_ context.Context, state *K6LoadTestRunState) (*action_kit_api.StopResult, error) {
	return stop(state)
}