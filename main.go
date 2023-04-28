/*
 * Copyright 2023 steadybit GmbH. All rights reserved.
 */

package main

import (
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/action-kit/go/action_kit_sdk"
	"github.com/steadybit/extension-kit/extbuild"
	"github.com/steadybit/extension-kit/exthealth"
	"github.com/steadybit/extension-kit/exthttp"
	"github.com/steadybit/extension-kit/extlogging"
	"github.com/steadybit/extension-scaffold/config"
	"github.com/steadybit/extension-scaffold/extk6"
)

func main() {
	extlogging.InitZeroLog()
	extbuild.PrintBuildInformation()
	exthealth.SetReady(false)
	exthealth.StartProbes(8088)

	config.ParseConfiguration()
	config.ValidateConfiguration()

	exthttp.RegisterHttpHandler("/", exthttp.GetterAsHandler(getExtensionList))

	action_kit_sdk.RegisterAction(extk6.NewK6LoadTestRunAction())

	action_kit_sdk.InstallSignalHandler()

	exthealth.SetReady(true)

	exthttp.Listen(exthttp.ListenOpts{
		Port: 8087,
	})
}

type ExtensionListResponse struct {
	action_kit_api.ActionList `json:",inline"`
}

func getExtensionList() ExtensionListResponse {
	return ExtensionListResponse{
		ActionList: action_kit_sdk.GetActionList(),
	}
}
