// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2024 Steadybit GmbH

package extk6

import (
	"context"
	"fmt"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/steadybit/discovery-kit/go/discovery_kit_sdk"
	"github.com/steadybit/extension-k6/config"
	"github.com/steadybit/extension-kit/extbuild"
	"github.com/steadybit/extension-kit/extutil"
	"os"
)

type k6LocationDiscovery struct{}

var (
	_ discovery_kit_sdk.TargetDescriber = (*k6LocationDiscovery)(nil)
)

func NewDiscovery() discovery_kit_sdk.TargetDiscovery {
	discovery := &k6LocationDiscovery{}
	return discovery_kit_sdk.NewCachedTargetDiscovery(discovery,
		discovery_kit_sdk.WithRefreshTargetsNow(),
		//No interval, target is not changing during runtime
	)
}

func (e *k6LocationDiscovery) Describe() discovery_kit_api.DiscoveryDescription {
	return discovery_kit_api.DiscoveryDescription{
		Id: targetType,
		Discover: discovery_kit_api.DescribingEndpointReferenceWithCallInterval{
			CallInterval: extutil.Ptr(fmt.Sprintf("%ds", 300)),
		},
	}
}

func (e *k6LocationDiscovery) DescribeTarget() discovery_kit_api.TargetDescription {
	return discovery_kit_api.TargetDescription{
		Id:       targetType,
		Label:    discovery_kit_api.PluralLabel{One: "K6 Location", Other: "K6 Locations"},
		Category: extutil.Ptr("execution locations"),
		Version:  extbuild.GetSemverVersionStringOrUnknown(),
		Icon:     extutil.Ptr(targetIcon),

		Table: discovery_kit_api.Table{
			Columns: []discovery_kit_api.Column{
				{Attribute: "k8s.cluster-name"},
				{Attribute: "k8s.namespace"},
				{Attribute: "aws.account", FallbackAttributes: &[]string{"gcp.project.id", "azure.subscription.id"}},
				{Attribute: "aws.zone", FallbackAttributes: &[]string{"gcp.zone", "azure.zone"}},
			},
			OrderBy: []discovery_kit_api.OrderBy{
				{
					Attribute: "k8s.cluster-name",
					Direction: "ASC",
				},
			},
		},
	}
}

func (e *k6LocationDiscovery) DiscoverTargets(_ context.Context) ([]discovery_kit_api.Target, error) {
	attributes := make(map[string][]string)

	var id, label string
	if (config.Config.KubernetesNamespace != "") && (config.Config.KubernetesPodName != "") && (config.Config.KubernetesNodeName != "") {
		id = fmt.Sprintf("%s-%s", config.Config.KubernetesNamespace, config.Config.KubernetesPodName)
		label = fmt.Sprintf("%s/%s", config.Config.KubernetesNamespace, config.Config.KubernetesPodName)
		attributes["k8s.namespace"] = []string{config.Config.KubernetesNamespace}
		attributes["k8s.pod.name"] = []string{config.Config.KubernetesPodName}
		attributes["k8s.node.name"] = []string{config.Config.KubernetesNodeName}
		attributes["host.hostname"] = []string{config.Config.KubernetesNodeName}
	} else {
		hostname, _ := os.Hostname()
		pid := os.Getpid()
		id = fmt.Sprintf("%s-%d", hostname, pid)
		label = fmt.Sprintf("%s/%d", hostname, pid)
		attributes["host.hostname"] = []string{hostname}
		attributes["process.pid"] = []string{fmt.Sprintf("%d", pid)}
	}

	if config.Config.KubernetesClusterName != "" {
		attributes["k8s.cluster-name"] = []string{config.Config.KubernetesClusterName}
	}

	return []discovery_kit_api.Target{
		{
			Id:         id,
			Label:      label,
			TargetType: targetType,
			Attributes: attributes,
		},
	}, nil
}
