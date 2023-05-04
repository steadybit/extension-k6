package extk6

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
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

const (
	actionIdPrefix = "com.github.steadybit.extension_k6"
	targetIcon     = "data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjEiIGhlaWdodD0iMjAiIHZpZXdCb3g9IjAgMCAyMSAyMCIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cGF0aCBkPSJNMTkuNjY2IDE4LjMzNEgxLjMzM0w3LjQzNiA1LjUyOWwzLjY3NyAyLjY1OEwxNS45MDguODMzbDMuNzU4IDE3LjV6bS02LjcyMi0yLjc2NmguMDRjLjQ1MyAwIC44OS0uMTcyIDEuMjE3LS40ODFhMS41NjggMS41NjggMCAwMC41MjMtMS4xODQgMS40MTcgMS40MTcgMCAwMC0uNTA0LTEuMTM2IDEuNTQ2IDEuNTQ2IDAgMDAtMS4wNDctLjQ0M2gtLjAzYS41NTIuNTUyIDAgMDAtLjE1Mi4wMmwuOTY4LTEuNDE0LS43NzEtLjUzLS4zNjUuNTMtLjkzMyAxLjRjLS4xNi4yMzMtLjI5NC40MzctLjM3Ny41OC0uMDg2LjE1LS4xNi4zMDctLjIyMi40NjgtLjA3LjE3Mi0uMTA2LjM1Ni0uMTA2LjU0YTEuNTQ1IDEuNTQ1IDAgMDAuNTE3IDEuMTcxYy4zMjMuMzEuNzU1LjQ4MiAxLjIwNi40ODFsLjAzNi0uMDAyem0tNC4wOTgtMS41MjNsMS4wNjggMS40ODZoMS4xNDNMOS44IDEzLjgwN2wxLjExNi0xLjUyNS0uNzQxLS41MDQtLjMyNy40MjUtMS4wMDQgMS4zOTJ2LTIuOGwtMS0uODAxdjUuNTM3aDF2LTEuNDg4bC4wMDIuMDAyem00LjEuNTk1YS43NDIuNzQyIDAgMDEtLjUyLS4yMTIuNzE3LjcxNyAwIDAxMC0xLjAyNC43NDIuNzQyIDAgMDEuNTItLjIxMWguMDA2YS43MjkuNzI5IDAgMDEuNTE5LjIxOC42NzYuNjc2IDAgMDEuMjIyLjUwMS43My43MyAwIDAxLS4yMjIuNTE0Ljc1NC43NTQgMCAwMS0uNTI1LjIxMnYuMDAyeiIgZmlsbD0iY3VycmVudENvbG9yIi8+PC9zdmc+"
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

func getActionDescription(actionId string, label string, description string) *action_kit_api.ActionDescription {
	return &action_kit_api.ActionDescription{
		Id:          actionId,
		Label:       label,
		Description: description,
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

func prepare(state *K6LoadTestRunState, request action_kit_api.PrepareActionRequestBody, command []string) (*action_kit_api.PrepareResult, error) {
	var config K6LoadTestRunConfig
	if err := extconversion.Convert(request.Config, &config); err != nil {
		return nil, extension_kit.ToError("Failed to unmarshal the config.", err)
	}

	state.ExecutionId = request.ExecutionId
	state.Timestamp = time.Now().Format(time.RFC3339)
	state.Command = command

	if config.Environment != nil {
		for _, value := range config.Environment {
			state.Command = append(state.Command, "--env")
			state.Command = append(state.Command, fmt.Sprintf("%s=%s", value["key"], value["value"]))
		}
	}

	return nil, nil
}

func start(state *K6LoadTestRunState) (*action_kit_api.StartResult, error) {
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

func status(state *K6LoadTestRunState) (*action_kit_api.StatusResult, error) {
	log.Debug().Msgf("Checking K6 status for %d\n", state.Pid)

	cmdState, err := extcmd.GetCmdState(state.CmdStateID)
	if err != nil {
		return nil, extension_kit.ToError("Failed to find command state", err)
	}

	var result action_kit_api.StatusResult

	// check if k6 is still running
	exitCode := cmdState.Cmd.ProcessState.ExitCode()
	stdOut := cmdState.GetLines(false)
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
		title := fmt.Sprintf("K6 run failed, exit-code %d", exitCode)
		stdOutError := extractErrorFromStdOut(stdOut)
		if stdOutError != nil {
			title = *stdOutError
		}
		result.Completed = true
		result.Error = &action_kit_api.ActionKitError{
			Status: extutil.Ptr(action_kit_api.Errored),
			Title:  title,
		}
	}

	filename := fmt.Sprintf("/tmp/steadybit/%v/k6_log.txt", state.ExecutionId) //Folder is managed by action_kit_sdk's file download handling
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

func extractErrorFromStdOut(lines []string) *string {
	//Find error, last log lines first
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.Contains(lines[i], "level=error") {
			split := strings.SplitAfter(lines[i], "msg=")
			if len(split) > 1 {
				return &split[1]
			}
		}
	}
	return nil
}

func stop(state *K6LoadTestRunState) (*action_kit_api.StopResult, error) {
	cmdState, err := extcmd.GetCmdState(state.CmdStateID)
	if err != nil {
		return nil, extension_kit.ToError("Failed to find command state", err)
	}
	extcmd.RemoveCmdState(state.CmdStateID)

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
