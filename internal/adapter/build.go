package adapter

import (
	"fmt"

	"github.com/akhmanov/agentspec/internal/adapter/claudecode"
	"github.com/akhmanov/agentspec/internal/adapter/opencode"
	"github.com/akhmanov/agentspec/internal/model"
)

var supportedTargets = []string{"opencode", "claude-code"}

func Build(target string, res *model.Resolved) (*model.Desired, error) {
	switch target {
	case "opencode":
		return opencode.Build(res), nil
	case "claude-code":
		return claudecode.Build(res), nil
	default:
		return nil, fmt.Errorf("unsupported target %q", target)
	}
}

func SupportedTarget(target string) bool {
	for _, supported := range supportedTargets {
		if target == supported {
			return true
		}
	}
	return false
}

func SupportedTargets() []string {
	return append([]string(nil), supportedTargets...)
}
