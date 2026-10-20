/*
 * Copyright 2024 steadybit GmbH. All rights reserved.
 */

package config

import (
	"strings"

	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog/log"
)

// Specification is the configuration specification for the extension. Configuration values can be applied
// through environment variables. Learn more through the documentation of the envconfig package.
// https://github.com/kelseyhightower/envconfig
type Specification struct {
	KubernetesClusterName   string `json:"kubernetesClusterName" split_words:"true" required:"false"`
	KubernetesNodeName      string `json:"kubernetesNodeName" split_words:"true" required:"false"`
	KubernetesPodName       string `json:"kubernetesPodName" split_words:"true" required:"false"`
	KubernetesNamespace     string `json:"kubernetesNamespace" split_words:"true" required:"false"`
	EnableLocationSelection bool   `json:"enableLocationSelection" split_words:"true" required:"false"`
	CloudApiToken           string `json:"cloudApiToken" split_words:"true" required:"false"`
	CloudApiBaseUrl         string `json:"CloudApiBaseUrl" split_words:"true" required:"false" default:"https://api.k6.io"`
	// K6EnvAllowList are additional variables of the extension's environment passed to k6, on top of the built-in ones.
	K6EnvAllowList []string `json:"k6EnvAllowList" split_words:"true" required:"false"`
}

var (
	Config Specification
)

func ParseConfiguration() {
	err := envconfig.Process("steadybit_extension", &Config)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to parse configuration from environment.")
	}
}

func ValidateConfiguration() {
	Config.K6EnvAllowList = sanitizeK6EnvAllowList(Config.K6EnvAllowList)
}

// sanitizeK6EnvAllowList drops the entries that must not or can't be passed to k6: the extension's own configuration
// and secrets (STEADYBIT_EXTENSION_*), and patterns, as only exact names are supported.
func sanitizeK6EnvAllowList(entries []string) []string {
	var allowed []string
	for _, entry := range entries {
		name := strings.TrimSpace(entry)
		switch {
		case name == "":
		case strings.HasPrefix(strings.ToUpper(name), "STEADYBIT_EXTENSION_"):
			log.Warn().Msgf("Ignoring %s in STEADYBIT_EXTENSION_K6_ENV_ALLOW_LIST: the extension's own configuration is never passed to k6.", name)
		case strings.Contains(name, "*"):
			log.Warn().Msgf("Ignoring %s in STEADYBIT_EXTENSION_K6_ENV_ALLOW_LIST: only exact variable names are supported.", name)
		default:
			allowed = append(allowed, name)
		}
	}
	return allowed
}
