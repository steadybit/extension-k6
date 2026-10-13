// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 Steadybit GmbH

package extk6

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/steadybit/extension-k6/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_discovery_Describe(t *testing.T) {
	d := &k6LocationDiscovery{}
	assert.Equal(t, targetType, d.Describe().Id)
	assert.Equal(t, "300s", *d.Describe().Discover.CallInterval)

	target := d.DescribeTarget()
	assert.Equal(t, targetType, target.Id)
	assert.Equal(t, "K6 Location", target.Label.One)
	assert.NotEmpty(t, target.Table.Columns)
}

func Test_discovery_DiscoverTargets(t *testing.T) {
	previous := config.Config
	t.Cleanup(func() { config.Config = previous })

	t.Run("kubernetes", func(t *testing.T) {
		config.Config = config.Specification{
			KubernetesClusterName: "dev-cluster",
			KubernetesNamespace:   "steadybit",
			KubernetesPodName:     "extension-k6-abc",
			KubernetesNodeName:    "node-1",
		}

		targets, err := (&k6LocationDiscovery{}).DiscoverTargets(context.Background())
		require.NoError(t, err)
		require.Len(t, targets, 1)
		assert.Equal(t, "steadybit-extension-k6-abc", targets[0].Id)
		assert.Equal(t, "steadybit/extension-k6-abc", targets[0].Label)
		assert.Equal(t, targetType, targets[0].TargetType)
		assert.Equal(t, map[string][]string{
			"k8s.cluster-name": {"dev-cluster"},
			"k8s.namespace":    {"steadybit"},
			"k8s.pod.name":     {"extension-k6-abc"},
			"k8s.node.name":    {"node-1"},
			"host.hostname":    {"node-1"},
		}, targets[0].Attributes)
	})

	t.Run("host", func(t *testing.T) {
		config.Config = config.Specification{}
		hostname, _ := os.Hostname()
		pid := os.Getpid()

		targets, err := (&k6LocationDiscovery{}).DiscoverTargets(context.Background())
		require.NoError(t, err)
		require.Len(t, targets, 1)
		assert.Equal(t, fmt.Sprintf("%s-%d", hostname, pid), targets[0].Id)
		assert.Equal(t, fmt.Sprintf("%s/%d", hostname, pid), targets[0].Label)
		assert.Equal(t, map[string][]string{
			"host.hostname": {hostname},
			"process.pid":   {fmt.Sprintf("%d", pid)},
		}, targets[0].Attributes)
	})
}

func Test_NewDiscovery(t *testing.T) {
	assert.NotNil(t, NewDiscovery())
}
