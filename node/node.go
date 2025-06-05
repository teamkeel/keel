package node

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/proto"
)

// IsEnabled returns true if the provided dir contains any tests or if the provided
// schema contains any functions.
func IsEnabled(dir string, s *proto.Schema, cfg *config.ProjectConfig) bool {
	return HasFunctions(s, cfg) || HasTests(dir)
}

// HasFunctions returns true if the schema contains any custom functions or jobs.
func HasFunctions(sch *proto.Schema, cfg *config.ProjectConfig) bool {
	var actions []*proto.Action

	for _, model := range sch.GetModels() {
		actions = append(actions, model.GetActions()...)
	}

	hasCustomFunctions := lo.SomeBy(actions, func(o *proto.Action) bool {
		return o.GetImplementation() == proto.ActionImplementation_ACTION_IMPLEMENTATION_CUSTOM
	})

	hasHooks := len(cfg.Auth.EnabledHooks()) > 0

	hasJobs := len(sch.GetJobs()) > 0

	hasSubscribers := len(sch.GetSubscribers()) > 0

	hasRoutes := len(sch.GetRoutes()) > 0

	hasFlows := len(sch.GetFlows()) > 0

	return hasCustomFunctions || hasHooks || hasJobs || hasSubscribers || hasRoutes || hasFlows
}

// HasTests returns true if there any TypeScript test files in dir or any of it's
// subdirectories.
func HasTests(dir string) bool {
	fs := os.DirFS(dir)

	// the only potential error returned from glob here is bad pattern,
	// which we know not to be true
	testFiles, _ := doublestar.Glob(fs, "**/*.test.ts")

	// there could be other *.test.ts files unrelated to the Keel testing framework,
	// so for each test, we do a naive check that the file contents includes a match
	// for the string "@teamkeel/testing"
	return lo.SomeBy(testFiles, func(path string) bool {
		b, err := os.ReadFile(filepath.Join(dir, path))

		if err != nil {
			return false
		}

		// todo: improve this check as its pretty naive
		return strings.Contains(string(b), "@teamkeel/testing")
	})
}
