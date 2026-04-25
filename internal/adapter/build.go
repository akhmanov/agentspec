package adapter

import (
	"fmt"

	"agentspec/internal/adapter/claudecode"
	"agentspec/internal/adapter/opencode"
	"agentspec/internal/model"
)

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
	_, err := Build(target, &model.Resolved{})
	return err == nil
}
