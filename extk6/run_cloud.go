/*
 * Copyright 2023 steadybit GmbH. All rights reserved.
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
	return *getActionDescription(fmt.Sprintf("%s.cloud", actionIdPrefix), "K6 Cloud", "Execute a K6 load using K6 Cloud.", nil)
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
	loggableToken := strings.Repeat("*", len(config.Config.CloudApiToken)-5) + config.Config.CloudApiToken[len(config.Config.CloudApiToken)-5:]
	log.Info().Msg("Use K6 cloud with token: " + loggableToken)
	return start(state, config.Config.CloudApiToken)
}

func (l *K6LoadTestCloudAction) Status(_ context.Context, state *K6LoadTestRunState) (*action_kit_api.StatusResult, error) {
	return status(state)
}

func (l *K6LoadTestCloudAction) Stop(_ context.Context, state *K6LoadTestRunState) (*action_kit_api.StopResult, error) {
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
	res, err := http.Get(fmt.Sprintf("https://api.k6.io/loadtests/v2/runs/%s", cloudRunId))
	if err != nil {
		return false, extension_kit.ToError("Failed to read k6 cloud status.", err)
	}
	defer res.Body.Close()

	var status StatusResponse
	if err := json.NewDecoder(res.Body).Decode(&status); err != nil {
		log.Error().Msgf("Failed to parse k6 cloud status: %s", err.Error())
		return false, extension_kit.ToError("Failed to parse k6 cloud status.", err)
	}

	return status.K6Run.RunStatus < 3, nil
}

func stopCloudRun(cloudRunId string) error {
	posturl := fmt.Sprintf("https://api.k6.io/loadtests/v2/runs/%s/stop", cloudRunId)

	// JSON body
	body := []byte(`{}`)

	// Create a HTTP post request
	r, err := http.NewRequest("POST", posturl, bytes.NewBuffer(body))
	if err != nil {
		return extension_kit.ToError("Failed to create post request to stop k6 cloud.", err)
	}
	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("Authorization", fmt.Sprintf("token %s", config.Config.CloudApiToken))

	client := &http.Client{}
	log.Info().Msgf("Stop K6 cloud at %s", posturl)
	res, err := client.Do(r)
	if err != nil {
		return extension_kit.ToError("Failed to stop k6 cloud.", err)
	}

	defer res.Body.Close()
	if res.StatusCode == http.StatusOK {
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
