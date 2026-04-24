#!/bin/sh
set -eu

ROOT=$(git rev-parse --show-toplevel)
cd "$ROOT"

npx -y @fission-ai/openspec@1.3.1 update --force

rm -f \
  .opencode/commands/build.md \
  .opencode/commands/debate.md \
  .opencode/commands/explore.md \
  .opencode/commands/plan.md \
  .opencode/commands/review.md \
  .opencode/commands/ship.md \
  .opencode/commands/spec.md \
  .opencode/commands/test.md \
  .opencode/commands/opsx-*.md \
  .opencode/agents/debate-architect.md \
  .opencode/agents/debate-moderator.md \
  .opencode/agents/debate-operator.md \
  .opencode/agents/debate-security.md \
  .opencode/agents/debate-skeptic.md \
  .opencode/agents/debate-user-advocate.md

rm -rf \
  .opencode/skills/openspec-* \
  .opencode/skills/workspace-build \
  .opencode/skills/workspace-debate \
  .opencode/skills/workspace-explore \
  .opencode/skills/workspace-planning \
  .opencode/skills/workspace-review \
  .opencode/skills/workspace-ship \
  .opencode/skills/workspace-spec \
  .opencode/skills/workspace-test

install -d \
  .opencode/commands \
  .opencode/instructions \
  .opencode/skills

cp openspec/opencode-overrides/opencode.json .opencode/opencode.json
cp openspec/opencode-overrides/instructions/INSTRUCTIONS.md .opencode/instructions/INSTRUCTIONS.md

cp openspec/opencode-overrides/commands/opsx-explore.md .opencode/commands/opsx-explore.md
cp openspec/opencode-overrides/commands/opsx-propose.md .opencode/commands/opsx-propose.md
cp openspec/opencode-overrides/commands/opsx-continue.md .opencode/commands/opsx-continue.md
cp openspec/opencode-overrides/commands/opsx-apply.md .opencode/commands/opsx-apply.md
cp openspec/opencode-overrides/commands/opsx-archive.md .opencode/commands/opsx-archive.md

cp -R openspec/opencode-overrides/skills/openspec-explore .opencode/skills/
cp -R openspec/opencode-overrides/skills/openspec-propose .opencode/skills/
cp -R openspec/opencode-overrides/skills/openspec-continue-change .opencode/skills/
cp -R openspec/opencode-overrides/skills/openspec-apply-change .opencode/skills/
cp -R openspec/opencode-overrides/skills/openspec-archive-change .opencode/skills/
