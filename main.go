package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strings"

	"github.com/akhmanov/agentspec/internal/adapter"
	"github.com/akhmanov/agentspec/internal/config"
	"github.com/akhmanov/agentspec/internal/model"
	"github.com/akhmanov/agentspec/internal/resolve"
	awsync "github.com/akhmanov/agentspec/internal/sync"
	"github.com/urfave/cli/v3"
)

const defaultConfigPath = "agentspec.yaml"
const defaultVersion = "dev"

var version = defaultVersion

var pseudoVersionPattern = regexp.MustCompile(`^v\d+\.\d+\.\d+(?:-|\-0\.)\d{14}-[0-9a-f]{12,}(?:\+dirty)?$`)

const (
	envRootPath   = "AGENTSPEC_ROOT"
	envConfigPath = "AGENTSPEC_CONFIG"
)

type executionContext struct {
	root       string
	configPath string
}

func main() {
	if err := newCommand().Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newCommand() *cli.Command {
	cli.VersionPrinter = func(cmd *cli.Command) {
		fmt.Fprintf(cmd.Root().Writer, "%s %s\n", cmd.Root().Name, cmd.Root().Version)
	}

	return &cli.Command{
		Name:    "agentspec",
		Usage:   "materialize workspace resources from agentspec.yaml",
		Version: reportedVersion(version, moduleVersion()),
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "root", Usage: "workspace root; falls back to AGENTSPEC_ROOT, then cwd"},
			&cli.StringFlag{Name: "config", Usage: "config path; falls back to AGENTSPEC_CONFIG, then <root>/agentspec.yaml"},
		},
		Commands: []*cli.Command{
			{
				Name:  "init",
				Usage: "write a starter agentspec.yaml",
				Action: func(_ context.Context, cmd *cli.Command) error {
					ctx, err := selectedExecutionContext(cmd)
					if err != nil {
						return err
					}
					return config.WriteStarter(ctx.configPath)
				},
			},
			{
				Name:  "plan",
				Usage: "preview workspace changes for one or more targets",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{Name: "target", Usage: targetFlagUsage("preview")},
					&cli.BoolFlag{Name: "verbose", Usage: "show one path per line and include conflict reasons"},
				},
				Action: func(_ context.Context, cmd *cli.Command) error {
					return runPlan(cmd)
				},
			},
			{
				Name:  "apply",
				Usage: "apply workspace changes for one or more targets",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{Name: "target", Usage: targetFlagUsage("apply")},
				},
				Action: func(_ context.Context, cmd *cli.Command) error {
					return runApply(cmd)
				},
			},
		},
	}
}

func reportedVersion(buildVersion, moduleVer string) string {
	if buildVersion != "" && buildVersion != defaultVersion && buildVersion != "(devel)" {
		return strings.TrimPrefix(buildVersion, "v")
	}
	if moduleVer != "" && moduleVer != "(devel)" && !pseudoVersionPattern.MatchString(moduleVer) {
		return strings.TrimPrefix(moduleVer, "v")
	}
	return defaultVersion
}

func moduleVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	return info.Main.Version
}

func runPlan(cmd *cli.Command) error {
	ctx, err := selectedExecutionContext(cmd)
	if err != nil {
		return err
	}

	targets, err := selectedTargets(cmd)
	if err != nil {
		return err
	}

	for i, target := range targets {
		_, des, err := loadDesired(ctx, target)
		if err != nil {
			return err
		}

		plan, err := awsync.Preview(ctx.root, target, des)
		if err != nil {
			return err
		}

		if i > 0 {
			fmt.Println()
		}
		if len(targets) > 1 {
			fmt.Printf("%s:\n", target)
		}
		printPlan(plan, cmd.Bool("verbose"))
	}
	return nil
}

func runApply(cmd *cli.Command) error {
	ctx, err := selectedExecutionContext(cmd)
	if err != nil {
		return err
	}

	targets, err := selectedTargets(cmd)
	if err != nil {
		return err
	}

	for _, target := range targets {
		_, des, err := loadDesired(ctx, target)
		if err != nil {
			return err
		}
		if err := awsync.Apply(ctx.root, target, des); err != nil {
			return err
		}
	}

	return nil
}

func selectedExecutionContext(cmd *cli.Command) (executionContext, error) {
	wd, err := os.Getwd()
	if err != nil {
		return executionContext{}, err
	}

	root := cmd.String("root")
	rootFromFlag := root != ""
	if root == "" {
		root = os.Getenv(envRootPath)
	}
	if root == "" {
		root = wd
	} else {
		root = cliPath(wd, root)
	}

	configPath := cmd.String("config")
	configFromFlag := configPath != ""
	if configPath == "" {
		configPath = os.Getenv(envConfigPath)
	}
	if configPath == "" {
		configPath = filepath.Join(root, defaultConfigPath)
	} else {
		configBase := root
		if configFromFlag && !rootFromFlag {
			configBase = wd
		}
		if rootFromFlag {
			configBase = root
		}
		configPath = cliPath(configBase, configPath)
	}

	return executionContext{root: root, configPath: configPath}, nil
}

func selectedTargets(cmd *cli.Command) ([]string, error) {
	rawTargets := cmd.StringSlice("target")
	if len(rawTargets) == 0 {
		return nil, fmt.Errorf("command requires a target flag")
	}

	targets := make([]string, 0, len(rawTargets))
	seen := make(map[string]struct{}, len(rawTargets))
	for _, target := range rawTargets {
		if !adapter.SupportedTarget(target) {
			return nil, fmt.Errorf("unsupported target %q (supported targets: %s)", target, strings.Join(adapter.SupportedTargets(), ", "))
		}
		if _, ok := seen[target]; ok {
			continue
		}
		seen[target] = struct{}{}
		targets = append(targets, target)
	}
	return targets, nil
}

func targetFlagUsage(action string) string {
	return fmt.Sprintf("target to %s; repeatable; supported: %s", action, strings.Join(adapter.SupportedTargets(), ", "))
}

func loadDesired(ctx executionContext, target string) (string, *model.Desired, error) {
	cfg, err := config.Load(ctx.configPath)
	if err != nil {
		return "", nil, err
	}

	res, err := resolve.Resolve(ctx.root, cfg)
	if err != nil {
		return "", nil, err
	}

	des, err := adapter.Build(target, res)
	if err != nil {
		return "", nil, err
	}

	return ctx.root, des, nil
}

func cliPath(base, raw string) string {
	if filepath.IsAbs(raw) {
		return raw
	}

	return filepath.Join(base, raw)
}

func printPlan(plan *awsync.Plan, verbose bool) {
	if len(plan.Changes) == 0 && len(plan.Conflicts) == 0 {
		fmt.Println("No managed changes.")
		return
	}

	for _, kind := range []awsync.ChangeKind{awsync.Create, awsync.Update, awsync.Delete} {
		paths := []string{}
		for _, item := range plan.Changes {
			if item.Kind == kind {
				paths = append(paths, item.Path)
			}
		}
		if len(paths) == 0 {
			continue
		}

		if verbose {
			fmt.Printf("%s (%d):\n", kind, len(paths))
			for _, path := range paths {
				fmt.Printf("  - %s\n", path)
			}
			continue
		}

		fmt.Printf("%s (%d): %s\n", kind, len(paths), strings.Join(paths, ", "))
	}

	if len(plan.Conflicts) == 0 {
		return
	}

	if verbose {
		fmt.Printf("conflict (%d):\n", len(plan.Conflicts))
		for _, item := range plan.Conflicts {
			fmt.Printf("  - %s: %s\n", item.Path, item.Reason)
		}
		return
	}

	paths := make([]string, 0, len(plan.Conflicts))
	for _, item := range plan.Conflicts {
		paths = append(paths, item.Path)
	}
	fmt.Printf("conflict (%d): %s\n", len(plan.Conflicts), strings.Join(paths, ", "))
}
