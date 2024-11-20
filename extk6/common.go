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
)

const (
	targetType = "com.steadybit.extension_k6.location"
	targetIcon = "data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjQiIGhlaWdodD0iMjQiIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KPHBhdGggZmlsbC1ydWxlPSJldmVub2RkIiBjbGlwLXJ1bGU9ImV2ZW5vZGQiIGQ9Ik00LjUgOS4xQzQuNSA0LjkyIDcuODc1IDEuNSAxMiAxLjVDMTYuMTI1IDEuNSAxOS41IDQuOTIgMTkuNSA5LjFDMTkuNSAxMy45MTM1IDEzLjcyMjQgMTkuMzEyNCAxMi43NzcgMjAuMTk1OEMxMi43MTQ5IDIwLjI1MzkgMTIuNjczNiAyMC4yOTI0IDEyLjY1NjIgMjAuMzFDMTIuNDY4OCAyMC40MDUgMTIuMTg3NSAyMC41IDEyIDIwLjVDMTEuODEyNSAyMC41IDExLjUzMTIgMjAuNDA1IDExLjM0MzggMjAuMzFDMTEuMzI2NCAyMC4yOTI0IDExLjI4NTEgMjAuMjUzOSAxMS4yMjMgMjAuMTk1OEMxMC4yNzc2IDE5LjMxMjQgNC41IDEzLjkxMzUgNC41IDkuMVpNNi4zNzUgOS4xQzYuMzc1IDEyLjMzIDEwLjAzMTIgMTYuNDE1IDEyIDE4LjMxNUMxMy45Njg4IDE2LjQxNSAxNy42MjUgMTIuMjM1IDE3LjYyNSA5LjFDMTcuNjI1IDUuOTY1IDE1LjA5MzggMy40IDEyIDMuNEM4LjkwNjI1IDMuNCA2LjM3NSA1Ljk2NSA2LjM3NSA5LjFaTTguMjUgOS4xQzguMjUgNy4wMSA5LjkzNzUgNS4zIDEyIDUuM0MxNC4wNjI1IDUuMyAxNS43NSA3LjAxIDE1Ljc1IDkuMUMxNS43NSAxMS4xOSAxNC4wNjI1IDEyLjkgMTIgMTIuOUM5LjkzNzUgMTIuOSA4LjI1IDExLjE5IDguMjUgOS4xWk0xMC4xMjUgOS4xQzEwLjEyNSAxMC4xNDUgMTAuOTY4OCAxMSAxMiAxMUMxMy4wMzEyIDExIDEzLjg3NSAxMC4xNDUgMTMuODc1IDkuMUMxMy44NzUgOC4wNTUgMTMuMDMxMiA3LjIgMTIgNy4yQzEwLjk2ODggNy4yIDEwLjEyNSA4LjA1NSAxMC4xMjUgOS4xWk01LjA3MzMyIDE2Ljg3NDVDNS41NTYyNyAxNi42MDY2IDUuNzMwNjEgMTUuOTk3OSA1LjQ2MjcgMTUuNTE0OUM1LjE5NDggMTUuMDMyIDQuNTg2MSAxNC44NTc2IDQuMTAzMTUgMTUuMTI1NUMyLjc0Mzg0IDE1Ljg3OTYgMiAxNi45NTE1IDIgMTguMjVDMiAxOS4xMTYxIDIuNDI1NTEgMTkuODUwNyAzLjAwNTQ4IDIwLjQyMjFDMy41ODE5MyAyMC45ODk5IDQuMzY0NzYgMjEuNDU1MyA1LjI1MTQyIDIxLjgyNDdDNy4wMjkxMyAyMi41NjU0IDkuNDE1NjkgMjMgMTIgMjNDMTQuNTg0MyAyMyAxNi45NzA5IDIyLjU2NTQgMTguNzQ4NiAyMS44MjQ3QzE5LjYzNTIgMjEuNDU1MyAyMC40MTgxIDIwLjk4OTkgMjAuOTk0NSAyMC40MjIxQzIxLjU3NDUgMTkuODUwNyAyMiAxOS4xMTYxIDIyIDE4LjI1QzIyIDE2Ljk1MTUgMjEuMjU2MiAxNS44Nzk2IDE5Ljg5NjkgMTUuMTI1NUMxOS40MTM5IDE0Ljg1NzYgMTguODA1MiAxNS4wMzIgMTguNTM3MyAxNS41MTQ5QzE4LjI2OTQgMTUuOTk3OSAxOC40NDM3IDE2LjYwNjYgMTguOTI2NyAxNi44NzQ1QzE5LjgyNzEgMTcuMzczOSAyMCAxNy44NjAxIDIwIDE4LjI1QzIwIDE4LjQxOTQgMTkuOTIxOCAxOC42NzEzIDE5LjU5MSAxOC45OTczQzE5LjI1NjUgMTkuMzI2NyAxOC43MjE0IDE5LjY2OTQgMTcuOTc5MyAxOS45Nzg2QzE2LjQ5OTcgMjAuNTk1MSAxNC4zODYzIDIxIDEyIDIxQzkuNjEzNzUgMjEgNy41MDAzMSAyMC41OTUxIDYuMDIwNjUgMTkuOTc4NkM1LjI3ODY0IDE5LjY2OTQgNC43NDM0NSAxOS4zMjY3IDQuNDA5MDUgMTguOTk3M0M0LjA3ODE3IDE4LjY3MTMgNCAxOC40MTk0IDQgMTguMjVDNCAxNy44NjAxIDQuMTcyOTUgMTcuMzczOSA1LjA3MzMyIDE2Ljg3NDVaIiBmaWxsPSIjMUQyNjMyIi8+Cjwvc3ZnPgo="

	actionIdPrefix = "com.steadybit.extension_k6"
	actionIcon     = "data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjEiIGhlaWdodD0iMjAiIHZpZXdCb3g9IjAgMCAyMSAyMCIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cGF0aCBkPSJNMTkuNjY2IDE4LjMzNEgxLjMzM0w3LjQzNiA1LjUyOWwzLjY3NyAyLjY1OEwxNS45MDguODMzbDMuNzU4IDE3LjV6bS02LjcyMi0yLjc2NmguMDRjLjQ1MyAwIC44OS0uMTcyIDEuMjE3LS40ODFhMS41NjggMS41NjggMCAwMC41MjMtMS4xODQgMS40MTcgMS40MTcgMCAwMC0uNTA0LTEuMTM2IDEuNTQ2IDEuNTQ2IDAgMDAtMS4wNDctLjQ0M2gtLjAzYS41NTIuNTUyIDAgMDAtLjE1Mi4wMmwuOTY4LTEuNDE0LS43NzEtLjUzLS4zNjUuNTMtLjkzMyAxLjRjLS4xNi4yMzMtLjI5NC40MzctLjM3Ny41OC0uMDg2LjE1LS4xNi4zMDctLjIyMi40NjgtLjA3LjE3Mi0uMTA2LjM1Ni0uMTA2LjU0YTEuNTQ1IDEuNTQ1IDAgMDAuNTE3IDEuMTcxYy4zMjMuMzEuNzU1LjQ4MiAxLjIwNi40ODFsLjAzNi0uMDAyem0tNC4wOTgtMS41MjNsMS4wNjggMS40ODZoMS4xNDNMOS44IDEzLjgwN2wxLjExNi0xLjUyNS0uNzQxLS41MDQtLjMyNy40MjUtMS4wMDQgMS4zOTJ2LTIuOGwtMS0uODAxdjUuNTM3aDF2LTEuNDg4bC4wMDIuMDAyem00LjEuNTk1YS43NDIuNzQyIDAgMDEtLjUyLS4yMTIuNzE3LjcxNyAwIDAxMC0xLjAyNC43NDIuNzQyIDAgMDEuNTItLjIxMWguMDA2YS43MjkuNzI5IDAgMDEuNTE5LjIxOC42NzYuNjc2IDAgMDEuMjIyLjUwMS43My43MyAwIDAxLS4yMjIuNTE0Ljc1NC43NTQgMCAwMS0uNTI1LjIxMnYuMDAyeiIgZmlsbD0iY3VycmVudENvbG9yIi8+PC9zdmc+"
)

type K6LoadTestRunState struct {
	Command     []string  `json:"command"`
	Pid         int       `json:"pid"`
	CmdStateID  string    `json:"cmdStateId"`
	ExecutionId uuid.UUID `json:"executionId"`
	CloudRunId  string    `json:"cloudRunId"`
}

type K6LoadTestRunConfig struct {
	Environment []map[string]string
	File        string
}

func getActionDescription(actionId string, label string, description string, hint *action_kit_api.ActionHint) *action_kit_api.ActionDescription {
	return &action_kit_api.ActionDescription{
		Id:          actionId,
		Label:       label,
		Description: description,
		Version:     extbuild.GetSemverVersionStringOrUnknown(),
		Icon:        extutil.Ptr(actionIcon),
		Technology:  extutil.Ptr("K6"),
		Kind:        action_kit_api.LoadTest,
		TimeControl: action_kit_api.TimeControlInternal,
		Hint:        hint,
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
				Order: extutil.Ptr(1),
			},
			{
				Name:        "environment",
				Label:       "Environment variables",
				Description: extutil.Ptr("Environment variables which will be accessible in your k6 script by ${__ENV.foobar}"),
				Type:        action_kit_api.KeyValue,
				Required:    extutil.Ptr(true),
				Order:       extutil.Ptr(2),
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
	state.Command = command

	if config.Environment != nil {
		for _, value := range config.Environment {
			state.Command = append(state.Command, "--env")
			state.Command = append(state.Command, fmt.Sprintf("%s=%s", value["key"], value["value"]))
		}
	}

	return nil, nil
}

func start(state *K6LoadTestRunState, token string) (*action_kit_api.StartResult, error) {
	log.Info().Msgf("Starting k6 load test with command: %s", strings.Join(state.Command, " "))
	cmd := exec.Command(state.Command[0], state.Command[1:]...)
	cmd.Env = os.Environ()
	if token != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("K6_CLOUD_TOKEN=%s", token))
	}
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
	addCloudRunIdToState(stdOut, state)
	stdOutToLog(stdOut)
	if exitCode == -1 {
		log.Debug().Msgf("K6 is still running")
		result.Completed = false
	} else if exitCode == 0 {
		log.Info().Msgf("K6 run completed successfully")
		result.Completed = true
	} else if exitCode == 97 || exitCode == 99 {
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

func addCloudRunIdToState(lines []string, state *K6LoadTestRunState) {
	for _, line := range lines {
		trimmed := strings.TrimSpace(strings.ReplaceAll(line, "\n", ""))
		cloudRunId := substringAfter(trimmed, "output: https://app.k6.io/runs/")
		if cloudRunId != nil {
			log.Info().Msgf("Found cloud run id: %s", *cloudRunId)
			state.CloudRunId = *cloudRunId
		}
	}
}

func substringAfter(value string, a string) *string {
	pos := strings.LastIndex(value, a)
	if pos == -1 {
		return nil
	}
	adjustedPos := pos + len(a)
	if adjustedPos >= len(value) {
		return nil
	}
	return extutil.Ptr(value[adjustedPos:])
}

func stdOutToLog(lines []string) {
	for _, line := range lines {
		trimmed := strings.TrimSpace(strings.ReplaceAll(line, "\n", ""))
		if len(trimmed) > 0 {
			log.Info().Msgf("---- %s", trimmed)
		}
	}
}

func stdOutToMessages(lines []string) []action_kit_api.Message {
	messages := make([]action_kit_api.Message, 0)
	for _, line := range lines {
		trimmed := strings.TrimSpace(strings.ReplaceAll(line, "\n", ""))
		if len(trimmed) > 0 {
			messages = append(messages, action_kit_api.Message{
				Level:   extutil.Ptr(action_kit_api.Info),
				Message: trimmed,
			})
		}
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
	if state.CmdStateID == "" {
		log.Info().Msg("K6 not yet started, nothing to stop.")
		return nil, nil
	}

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
	stdOutToLog(stdOut)
	filename := fmt.Sprintf("/tmp/steadybit/%v/k6_log.txt", state.ExecutionId) //Folder is managed by action_kit_sdk's file download handling
	if err := extfile.AppendToFile(filename, stdOut); err != nil {
		return nil, extension_kit.ToError("Failed to append log to file", err)
	}
	messages := stdOutToMessages(stdOut)

	// read return code and send it as Message
	exitCode := cmdState.Cmd.ProcessState.ExitCode()
	if exitCode != 0 && exitCode != -1 {
		messages = append(messages, action_kit_api.Message{
			Level:   extutil.Ptr(action_kit_api.Error),
			Message: fmt.Sprintf("K6 run failed with exit code %d", exitCode),
		})
	}

	artifacts := make([]action_kit_api.Artifact, 0)

	// check if log file exists and send it as artifact
	stats, err := os.Stat(filename)
	if err == nil { // file exists
		if stats.Size() > 1000000 {
			//zip if more than 1mb
			zippedLog := fmt.Sprintf("/tmp/steadybit/%v/k6_log.zip", state.ExecutionId)
			log.Info().Msgf("Zip log with command: %s %s %s", "zip", zippedLog, filename)
			zipCommand := exec.Command("zip", zippedLog, filename)
			zipErr := zipCommand.Run()
			if zipErr != nil {
				return nil, extension_kit.ToError("Failed to zip log", err)
			}
			content, err := extfile.File2Base64(zippedLog)
			if err != nil {
				return nil, err
			}
			artifacts = append(artifacts, action_kit_api.Artifact{
				Label: "$(experimentKey)_$(executionId)_k6_log.zip",
				Data:  content,
			})
		} else {
			content, err := extfile.File2Base64(filename)
			if err != nil {
				return nil, err
			}
			artifacts = append(artifacts, action_kit_api.Artifact{
				Label: "$(experimentKey)_$(executionId)_k6_log.txt",
				Data:  content,
			})
		}
	}

	metricsFilename := fmt.Sprintf("/tmp/steadybit/%v/metrics.json", state.ExecutionId)
	stats, err = os.Stat(metricsFilename)
	if err == nil { // file exists
		if stats.Size() > 1000000 {
			//zip if more than 1mb
			zippedMetrics := fmt.Sprintf("/tmp/steadybit/%v/metrics.zip", state.ExecutionId)
			log.Info().Msgf("Zip metrics with command: %s %s %s", "zip", zippedMetrics, metricsFilename)
			zipCommand := exec.Command("zip", zippedMetrics, metricsFilename)
			zipErr := zipCommand.Run()
			if zipErr != nil {
				return nil, extension_kit.ToError("Failed to zip metrics", err)
			}
			content, err := extfile.File2Base64(zippedMetrics)
			if err != nil {
				return nil, err
			}
			artifacts = append(artifacts, action_kit_api.Artifact{
				Label: "$(experimentKey)_$(executionId)_k6_metrics.zip",
				Data:  content,
			})
		} else {
			content, err := extfile.File2Base64(metricsFilename)
			if err != nil {
				return nil, err
			}
			artifacts = append(artifacts, action_kit_api.Artifact{
				Label: "$(experimentKey)_$(executionId)_k6_metrics.json",
				Data:  content,
			})
		}
	}

	log.Debug().Msgf("Returning %d messages", len(messages))
	return &action_kit_api.StopResult{
		Artifacts: extutil.Ptr(artifacts),
		Messages:  extutil.Ptr(messages),
	}, nil
}

func filter[T any](ss []T, test func(T) bool) (ret []T) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}
