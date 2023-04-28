/*
 * Copyright 2023 steadybit GmbH. All rights reserved.
 */

package extk6

import (
	"context"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/action-kit/go/action_kit_sdk"
)

type K6LoadTestRunAction struct{}

// Make sure action implements all required interfaces
var (
	_ action_kit_sdk.Action[K6LoadTestRunState]           = (*K6LoadTestRunAction)(nil)
	_ action_kit_sdk.ActionWithStatus[K6LoadTestRunState] = (*K6LoadTestRunAction)(nil)
	_ action_kit_sdk.ActionWithStop[K6LoadTestRunState]   = (*K6LoadTestRunAction)(nil)
)

type K6LoadTestRunState struct {
}

func NewK6LoadTestRunAction() action_kit_sdk.Action[K6LoadTestRunState] {
	return &K6LoadTestRunAction{}
}

func (l *K6LoadTestRunAction) NewEmptyState() K6LoadTestRunState {
	return K6LoadTestRunState{}
}

func (l *K6LoadTestRunAction) Describe() action_kit_api.ActionDescription {
	//TODO implement me
	panic("implement me")
}

func (l *K6LoadTestRunAction) Prepare(ctx context.Context, state *K6LoadTestRunState, request action_kit_api.PrepareActionRequestBody) (*action_kit_api.PrepareResult, error) {
	//TODO implement me
	panic("implement me")
}

func (l *K6LoadTestRunAction) Start(ctx context.Context, state *K6LoadTestRunState) (*action_kit_api.StartResult, error) {
	//TODO implement me
	panic("implement me")
}

func (l *K6LoadTestRunAction) Status(ctx context.Context, state *K6LoadTestRunState) (*action_kit_api.StatusResult, error) {
	//TODO implement me
	panic("implement me")
}

func (l *K6LoadTestRunAction) Stop(ctx context.Context, state *K6LoadTestRunState) (*action_kit_api.StopResult, error) {
	//TODO implement me
	panic("implement me")
}
