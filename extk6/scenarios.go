/*
 * Copyright 2026 steadybit GmbH. All rights reserved.
 */

package extk6

import (
	"context"
	"encoding/json"
	"os/exec"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	scenariosConflictTitle = "The VUs and Duration parameters can't be combined with k6 scenarios. Clear the parameters or remove the scenarios."
	inspectTimeout         = 30 * time.Second
)

// inspectK6 runs `k6 inspect` and returns its stdout, the options of the script or archive as JSON. A variable so
// tests can replace it.
var inspectK6 = func(ctx context.Context, args []string) ([]byte, error) {
	return exec.CommandContext(ctx, "k6", append([]string{"inspect"}, args...)...).Output()
}

// scriptDefinesScenarios reports whether the script (or archive) at file defines k6 scenarios, evaluating its options
// with the step's environment variables. If k6 can't inspect it, it returns false: the run itself then reports the
// actual problem (e.g. a syntax error).
func scriptDefinesScenarios(file string, envArgs []string) bool {
	args := append(append([]string{}, envArgs...), file)

	ctx, cancel := context.WithTimeout(context.Background(), inspectTimeout)
	defer cancel()
	out, err := inspectK6(ctx, args)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to inspect the k6 script, skipping the scenarios check.")
		return false
	}

	var options struct {
		Scenarios map[string]json.RawMessage `json:"scenarios"`
	}
	if err := json.Unmarshal(out, &options); err != nil {
		log.Warn().Err(err).Msg("Failed to parse the k6 inspect output, skipping the scenarios check.")
		return false
	}
	return len(options.Scenarios) > 0
}
