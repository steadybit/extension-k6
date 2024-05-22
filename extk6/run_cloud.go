/*
 * Copyright 2024 steadybit GmbH. All rights reserved.
 */

package extk6

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/action-kit/go/action_kit_sdk"
	"github.com/steadybit/extension-k6/config"
	extension_kit "github.com/steadybit/extension-kit"
	"github.com/steadybit/extension-kit/extconversion"
	"net/http"
	"strings"
)

type k6LoadTestCloudAction struct {
	baseUrl string
}

// Make sure action implements all required interfaces
var (
	_ action_kit_sdk.Action[K6LoadTestRunState]           = (*k6LoadTestCloudAction)(nil)
	_ action_kit_sdk.ActionWithStatus[K6LoadTestRunState] = (*k6LoadTestCloudAction)(nil)
	_ action_kit_sdk.ActionWithStop[K6LoadTestRunState]   = (*k6LoadTestCloudAction)(nil)
)

func NewK6LoadTestCloudAction() action_kit_sdk.Action[K6LoadTestRunState] {
	return &k6LoadTestCloudAction{}
}

func (l *k6LoadTestCloudAction) NewEmptyState() K6LoadTestRunState {
	return K6LoadTestRunState{}
}

func (l *k6LoadTestCloudAction) Describe() action_kit_api.ActionDescription {
	return *getActionDescription(fmt.Sprintf("%s.cloud", actionIdPrefix), "K6 Cloud", "Execute a K6 load using K6 Cloud.", nil)
}

func (l *k6LoadTestCloudAction) Prepare(_ context.Context, state *K6LoadTestRunState, request action_kit_api.PrepareActionRequestBody) (*action_kit_api.PrepareResult, error) {
	var runConfig K6LoadTestRunConfig
	if err := extconversion.Convert(request.Config, &runConfig); err != nil {
		return nil, extension_kit.ToError("Failed to unmarshal the runConfig.", err)
	}
	command := []string{"k6", "cloud", runConfig.File}
	return prepare(state, request, command)
}

func (l *k6LoadTestCloudAction) Start(_ context.Context, state *K6LoadTestRunState) (*action_kit_api.StartResult, error) {
	loggableToken := strings.Repeat("*", len(config.Config.CloudApiToken)-5) + config.Config.CloudApiToken[len(config.Config.CloudApiToken)-5:]
	log.Info().Msg("Use K6 cloud with token: " + loggableToken)
	return start(state, config.Config.CloudApiToken)
}

func (l *k6LoadTestCloudAction) Status(_ context.Context, state *K6LoadTestRunState) (*action_kit_api.StatusResult, error) {
	return status(state)
}

func (l *k6LoadTestCloudAction) Stop(_ context.Context, state *K6LoadTestRunState) (*action_kit_api.StopResult, error) {
	if state.CloudRunId != "" {
		running, err := isCloudRunStillRunning(state.CloudRunId)
		if err != nil {
			return nil, err
		}
		if running {
			err = stopCloudRun(state.CloudRunId)
			if err != nil {
				return nil, err
			}
		}
	}

	return stop(state)
}

func isCloudRunStillRunning(cloudRunId string) (bool, error) {
	res, err := http.Get(fmt.Sprintf("%s/loadtests/v2/runs/%s", config.Config.CloudApiBaseUrl, cloudRunId))
	if err != nil {
		return false, fmt.Errorf("failed to get k6 cloud status: %s", err.Error())
	} else if !(res.StatusCode >= http.StatusOK && res.StatusCode < http.StatusMultipleChoices) {
		return false, fmt.Errorf("failed to get k6 cloud status: %d - %s", res.StatusCode, res.Status)
	}
	defer func() { _ = res.Body.Close() }()

	var status StatusResponse
	if err := json.NewDecoder(res.Body).Decode(&status); err != nil {
		return false, fmt.Errorf("failed to get k6 cloud status: %s", err.Error())
	}

	return status.K6Run.RunStatus < 3, nil
}

func stopCloudRun(cloudRunId string) error {
	r, err := http.NewRequest("POST", fmt.Sprintf("%s/loadtests/v2/runs/%s/stop", config.Config.CloudApiBaseUrl, cloudRunId), bytes.NewBufferString("{}"))
	if err != nil {
		return fmt.Errorf("failed to stop k6 cloud: %w", err)
	}
	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("Authorization", fmt.Sprintf("token %s", config.Config.CloudApiToken))

	log.Info().Msgf("Stop K6 cloud at %s", r.URL)
	res, err := http.DefaultClient.Do(r)
	if err != nil {
		return fmt.Errorf("failed to stop k6 cloud: %w", err)
	}

	defer func() { _ = res.Body.Close() }()
	if res.StatusCode >= http.StatusOK && res.StatusCode < http.StatusMultipleChoices {
		log.Info().Msg("K6 cloud stop requested.")
	} else {
		log.Info().Msgf("K6 cloud stop responded with HTTP Status %s.", res.Status)
	}
	return nil
}

type StatusResponse struct {
	K6Run RunStatus `json:"k6-run"`
}

type RunStatus struct {
	// Values: https://k6.io/docs/cloud/cloud-reference/cloud-rest-api/test-runs/#read-load-test-run
	RunStatus int    `json:"run_status"`
	Started   string `json:"started"`
	Ended     string `json:"ended"`
}
