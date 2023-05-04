/*
 * Copyright 2023 steadybit GmbH. All rights reserved.
 */

package extk6

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/action-kit/go/action_kit_sdk"
	extension_kit "github.com/steadybit/extension-kit"
	"github.com/steadybit/extension-kit/extbuild"
	"github.com/steadybit/extension-kit/extcmd"
	"github.com/steadybit/extension-kit/extconversion"
	"github.com/steadybit/extension-kit/extfile"
	"github.com/steadybit/extension-kit/extutil"
	"os"
	"os/exec"
	"strings"
	"time"
)

type K6LoadTestRunAction struct{}

// Make sure action implements all required interfaces
var (
	_ action_kit_sdk.Action[K6LoadTestRunState]           = (*K6LoadTestRunAction)(nil)
	_ action_kit_sdk.ActionWithStatus[K6LoadTestRunState] = (*K6LoadTestRunAction)(nil)
	_ action_kit_sdk.ActionWithStop[K6LoadTestRunState]   = (*K6LoadTestRunAction)(nil)
)

type K6LoadTestRunState struct {
	Command         []string  `json:"command"`
	Pid             int       `json:"pid"`
	CmdStateID      string    `json:"cmdStateId"`
	Timestamp       string    `json:"timestamp"`
	StdOutLineCount int       `json:"stdOutLineCount"`
	ExecutionId     uuid.UUID `json:"executionId"`
}

type K6LoadTestRunConfig struct {
	Environment []map[string]string
	File        string
}

func NewK6LoadTestRunAction() action_kit_sdk.Action[K6LoadTestRunState] {
	return &K6LoadTestRunAction{}
}

func (l *K6LoadTestRunAction) NewEmptyState() K6LoadTestRunState {
	return K6LoadTestRunState{}
}

func (l *K6LoadTestRunAction) Describe() action_kit_api.ActionDescription {
	return action_kit_api.ActionDescription{
		Id:          fmt.Sprintf("%s.run", actionIdPrefix),
		Label:       "K6",
		Description: "Execute a K6 load test.",
		Version:     extbuild.GetSemverVersionStringOrUnknown(),
		Icon:        extutil.Ptr(targetIcon),
		Kind:        action_kit_api.LoadTest,
		TimeControl: action_kit_api.Internal,
		Parameters: []action_kit_api.ActionParameter{
			{
				Name:        "file",
				Label:       "K6 Script",
				Description: extutil.Ptr("Upload your K6 Script"),
				Type:        action_kit_api.File,
				Required:    extutil.Ptr(true),
				AcceptedFileTypes: extutil.Ptr([]string{
					".js",
				}),
			},
			{
				Name:        "environment",
				Label:       "Environment variables",
				Description: extutil.Ptr("Environment variables which will be accessible in your k6 script by ${__ENV.foobar}"),
				Type:        action_kit_api.KeyValue,
				Required:    extutil.Ptr(true),
			},
		},
		Status: extutil.Ptr(action_kit_api.MutatingEndpointReferenceWithCallInterval{
			CallInterval: extutil.Ptr("5s"),
		}),
		Stop: extutil.Ptr(action_kit_api.MutatingEndpointReference{}),
	}
}

func (l *K6LoadTestRunAction) Prepare(_ context.Context, state *K6LoadTestRunState, request action_kit_api.PrepareActionRequestBody) (*action_kit_api.PrepareResult, error) {
	var config K6LoadTestRunConfig
	if err := extconversion.Convert(request.Config, &config); err != nil {
		return nil, extension_kit.ToError("Failed to unmarshal the config.", err)
	}

	filename := fmt.Sprintf("/tmp/steadybit/%v/metrics.json", request.ExecutionId) //Folder is managed by action_kit_sdk's file download handling

	state.ExecutionId = request.ExecutionId
	state.Timestamp = time.Now().Format(time.RFC3339)
	state.Command = []string{
		"k6",
		"run",
		config.File,
		"--no-usage-report",
		"--out",
		fmt.Sprintf("json=%s", filename),
	}

	if config.Environment != nil {
		for _, value := range config.Environment {
			state.Command = append(state.Command, "--env")
			state.Command = append(state.Command, fmt.Sprintf("%s=%s", value["key"], value["value"]))
		}
	}

	return nil, nil
}

func (l *K6LoadTestRunAction) Start(_ context.Context, state *K6LoadTestRunState) (*action_kit_api.StartResult, error) {
	log.Info().Msgf("Starting k6 load test with command: %s", strings.Join(state.Command, " "))
	cmd := exec.Command(state.Command[0], state.Command[1:]...)
	cmdState := extcmd.NewCmdState(cmd)
	state.CmdStateID = cmdState.Id
	err := cmd.Start()
	if err != nil {
		return nil, extension_kit.ToError("Failed to start command.", err)
	}

	state.Pid = cmd.Process.Pid
	go func() {
		cmdErr := cmd.Wait()
		if cmdErr != nil {
			log.Error().Msgf("Failed to execute k6: %s", cmdErr)
		}
	}()
	log.Info().Msgf("Started load test.")

	state.Command = nil
	return nil, nil
}

func (l *K6LoadTestRunAction) Status(_ context.Context, state *K6LoadTestRunState) (*action_kit_api.StatusResult, error) {
	log.Debug().Msgf("Checking K6 status for %d\n", state.Pid)

	cmdState, err := extcmd.GetCmdState(state.CmdStateID)
	if err != nil {
		return nil, extension_kit.ToError("Failed to find command state", err)
	}

	var result action_kit_api.StatusResult

	// check if k6 is still running
	exitCode := cmdState.Cmd.ProcessState.ExitCode()
	if exitCode == -1 {
		log.Debug().Msgf("K6 is still running")
		result.Completed = false
	} else if exitCode == 0 {
		log.Info().Msgf("K6 run completed successfully")
		result.Completed = true
	} else if exitCode == 99 {
		log.Info().Msgf("K6 run completed with threshold failures")
		result.Completed = true
		result.Error = &action_kit_api.ActionKitError{
			Status: extutil.Ptr(action_kit_api.Failed),
			Title:  "Some thresholds have failed.",
		}
	} else {
		result.Completed = true
		result.Error = &action_kit_api.ActionKitError{
			Status: extutil.Ptr(action_kit_api.Errored),
			Title:  fmt.Sprintf("K6 run failed, exit-code %d", exitCode),
		}
	}

	filename := fmt.Sprintf("/tmp/steadybit/%v/k6_log.txt", state.ExecutionId) //Folder is managed by action_kit_sdk's file download handling
	stdOut := cmdState.GetLines(false)
	if err := extfile.AppendToFile(filename, stdOut); err != nil {
		return nil, extension_kit.ToError("Failed to append log to file", err)
	}
	messages := stdOutToMessages(stdOut)
	log.Debug().Msgf("Returning %d messages", len(messages))

	result.Messages = extutil.Ptr(messages)
	return &result, nil
}

func stdOutToMessages(lines []string) []action_kit_api.Message {
	var messages []action_kit_api.Message
	for _, line := range lines {
		messages = append(messages, action_kit_api.Message{
			Level:   extutil.Ptr(action_kit_api.Info),
			Message: line,
		})
	}
	return messages
}

func (l *K6LoadTestRunAction) Stop(_ context.Context, state *K6LoadTestRunState) (*action_kit_api.StopResult, error) {
	cmdState, err := extcmd.GetCmdState(state.CmdStateID)
	if err != nil {
		return nil, extension_kit.ToError("Failed to find command state", err)
	}

	// kill k6 if it is still running
	var pid = state.Pid
	process, err := os.FindProcess(pid)
	if err != nil {
		return nil, extension_kit.ToError("Failed to find process", err)
	}
	_ = process.Kill()

	// read Stout and Stderr and send it as Messages
	stdOut := cmdState.GetLines(true)
	filename := fmt.Sprintf("/tmp/steadybit/%v/k6_log.txt", state.ExecutionId) //Folder is managed by action_kit_sdk's file download handling
	if err := extfile.AppendToFile(filename, stdOut); err != nil {
		return nil, extension_kit.ToError("Failed to append log to file", err)
	}
	messages := stdOutToMessages(stdOut)

	// read return code and send it as Message
	exitCode := cmdState.Cmd.ProcessState.ExitCode()
	if exitCode != 0 {
		messages = append(messages, action_kit_api.Message{
			Level:   extutil.Ptr(action_kit_api.Error),
			Message: fmt.Sprintf("K6 run failed with exit code %d", exitCode),
		})
	}

	var artifacts []action_kit_api.Artifact

	// check if log file exists and send it as artifact
	_, err = os.Stat(filename)
	if err == nil { // file exists
		content, err := extfile.File2Base64(filename)
		if err != nil {
			return nil, err
		}
		artifacts = append(artifacts, action_kit_api.Artifact{
			Label: "$(experimentKey)_$(executionId)_k6_log.txt",
			Data:  content,
		})
	}

	metricsFilename := fmt.Sprintf("/tmp/steadybit/%v/metrics.json", state.ExecutionId)
	_, err = os.Stat(metricsFilename)
	if err == nil { // file exists
		content, err := extfile.File2Base64(metricsFilename)
		if err != nil {
			return nil, err
		}
		artifacts = append(artifacts, action_kit_api.Artifact{
			Label: "$(experimentKey)_$(executionId)_k6_metrics.json",
			Data:  content,
		})
	}

	log.Debug().Msgf("Returning %d messages", len(messages))
	return &action_kit_api.StopResult{
		Artifacts: extutil.Ptr(artifacts),
		Messages:  extutil.Ptr(messages),
	}, nil
}
