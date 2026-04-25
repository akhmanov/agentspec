package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"agentspec/internal/adapter"
	"agentspec/internal/config"
	"agentspec/internal/model"
	"agentspec/internal/resolve"
	awsync "agentspec/internal/sync"
	"github.com/urfave/cli/v3"
)

const defaultConfigPath = "agentspec.yaml"

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
	return &cli.Command{
		Name:  "agentspec",
		Usage: "materialize workspace resources from agentspec.yaml",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "root", Usage: "workspace root for config, outputs, and state"},
			&cli.StringFlag{Name: "config", Usage: "path to agentspec config file"},
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
					&cli.StringSliceFlag{Name: "target", Usage: "target workspace surface to preview"},
					&cli.BoolFlag{Name: "verbose"},
				},
				Action: func(_ context.Context, cmd *cli.Command) error {
					return runPlan(cmd)
				},
			},
			{
				Name:  "apply",
				Usage: "apply workspace changes for one or more targets",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{Name: "target", Usage: "target workspace surface to apply"},
				},
				Action: func(_ context.Context, cmd *cli.Command) error {
					return runApply(cmd)
				},
			},
		},
	}
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
	if root == "" {
		root = os.Getenv(envRootPath)
	}
	if root == "" {
		root = wd
	} else {
		root = cliPath(wd, root)
	}

	configPath := cmd.String("config")
	if configPath == "" {
		configPath = os.Getenv(envConfigPath)
	}
	if configPath == "" {
		configPath = filepath.Join(root, defaultConfigPath)
	} else {
		configPath = cliPath(root, configPath)
	}

	return executionContext{root: root, configPath: configPath}, nil
}


func selectedTargets(cmd *cli.Command) ([]string, error) {
	targets := cmd.StringSlice("target")
	if len(targets) == 0 {
		return nil, fmt.Errorf("command requires a target flag")
	}
	for _, target := range targets {
		if !adapter.SupportedTarget(target) {
			return nil, fmt.Errorf("unsupported target %q", target)
		}
	}
	return targets, nil
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
