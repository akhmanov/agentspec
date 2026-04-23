package main

import (
	"context"
	"fmt"
	"os"

	adapter "aw/internal/adapter/opencode"
	"aw/internal/config"
	"aw/internal/model"
	"aw/internal/resolve"
	awsync "aw/internal/sync"
	"github.com/urfave/cli/v3"
)

func main() {
	if err := newCommand().Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newCommand() *cli.Command {
	return &cli.Command{
		Name:  "aw",
		Usage: "materialize workspace resources from aw.yaml",
		Commands: []*cli.Command{
			{
				Name:  "init",
				Usage: "write a starter aw.yaml",
				Action: func(context.Context, *cli.Command) error {
					return config.WriteStarter("aw.yaml")
				},
			},
			{
				Name:  "plan",
				Usage: "preview workspace changes for a target",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "opencode"},
				},
				Action: func(_ context.Context, cmd *cli.Command) error {
					return runPlan(cmd)
				},
			},
			{
				Name:  "apply",
				Usage: "apply workspace changes for a target",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "opencode"},
				},
				Action: func(_ context.Context, cmd *cli.Command) error {
					return runApply(cmd)
				},
			},
		},
	}
}

func runPlan(cmd *cli.Command) error {
	target, err := selectedTarget(cmd)
	if err != nil {
		return err
	}

	root, des, err := loadDesired(target)
	if err != nil {
		return err
	}

	plan, err := awsync.Preview(root, target, des)
	if err != nil {
		return err
	}

	printPlan(plan)
	return nil
}

func runApply(cmd *cli.Command) error {
	target, err := selectedTarget(cmd)
	if err != nil {
		return err
	}

	root, des, err := loadDesired(target)
	if err != nil {
		return err
	}

	return awsync.Apply(root, target, des)
}

func selectedTarget(cmd *cli.Command) (string, error) {
	if cmd.Bool("opencode") {
		return "opencode", nil
	}

	return "", fmt.Errorf("command requires a target flag")
}

func loadDesired(target string) (string, *model.Desired, error) {
	root, err := os.Getwd()
	if err != nil {
		return "", nil, err
	}

	cfg, err := config.Load("aw.yaml")
	if err != nil {
		return "", nil, err
	}

	res, err := resolve.Resolve(root, cfg)
	if err != nil {
		return "", nil, err
	}

	if target != "opencode" {
		return "", nil, fmt.Errorf("unsupported target %q", target)
	}

	return root, adapter.Build(res), nil
}

func printPlan(plan *awsync.Plan) {
	if len(plan.Changes) == 0 && len(plan.Conflicts) == 0 {
		fmt.Println("No managed changes.")
		return
	}

	for _, item := range plan.Changes {
		fmt.Printf("%s %s\n", item.Kind, item.Path)
	}
	for _, item := range plan.Conflicts {
		fmt.Printf("conflict %s: %s\n", item.Path, item.Reason)
	}
}
