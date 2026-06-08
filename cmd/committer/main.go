package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/TangoEnSkai/committer-go/committer"
	"github.com/urfave/cli/v2"
)

const hookContent = `#!/bin/sh
committer "$1"
`

func main() {
	cfg, err := committer.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not load config: %v\n", err)
		cfg = committer.DefaultConfig()
	}

	app := &cli.App{
		Name:  "committer",
		Usage: "Conventional Commits message validator",
		// Default action: hook mode — committer <path-to-commit-msg-file>
		Action: func(c *cli.Context) error {
			if c.NArg() < 1 {
				return cli.Exit("unexpected error: no arguments passed to this script", 1)
			}
			return runHook(c.Args().First(), cfg)
		},
		Commands: []*cli.Command{
			{
				Name:      "lint",
				Usage:     "Validate a commit message string",
				ArgsUsage: "<message>",
				Action: func(c *cli.Context) error {
					if c.NArg() < 1 {
						return cli.Exit("lint requires a commit message argument", 1)
					}
					msg := committer.Message(c.Args().First())
					errMsg, ok := committer.ValidateMessage(msg, cfg)
					if !ok {
						committer.Print(msg, errMsg)
						return cli.Exit("", 1)
					}
					fmt.Println("commit message is valid")
					return nil
				},
			},
			{
				Name:  "install",
				Usage: "Install the commit-msg hook into the nearest .git directory",
				Action: func(c *cli.Context) error {
					return runInstall()
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		// cli.Exit errors are already printed by the framework; others are not
		os.Exit(1)
	}
}

// runHook is the original hook-mode logic.
func runHook(path string, cfg committer.Config) error {
	commitMsg, err := committer.Read(path)
	if err != nil {
		committer.Print(commitMsg, err.Error())
		return cli.Exit("", 1)
	}

	errMsg, ok := committer.ValidateMessage(commitMsg, cfg)
	if !ok {
		committer.Print(commitMsg, errMsg)
		return cli.Exit("", 1)
	}

	return nil
}

// runInstall writes the commit-msg hook to .git/hooks/commit-msg.
func runInstall() error {
	gitDir, err := findGitDir(".")
	if err != nil {
		return cli.Exit(fmt.Sprintf("could not find .git directory: %v", err), 1)
	}

	hooksDir := filepath.Join(gitDir, "hooks")
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		return cli.Exit(fmt.Sprintf("could not create hooks directory: %v", err), 1)
	}

	hookPath := filepath.Join(hooksDir, "commit-msg")
	if err := os.WriteFile(hookPath, []byte(hookContent), 0o755); err != nil {
		return cli.Exit(fmt.Sprintf("could not write hook file: %v", err), 1)
	}

	fmt.Printf("installed commit-msg hook at %s\n", hookPath)
	return nil
}

// findGitDir walks up from dir looking for a .git directory.
func findGitDir(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	for {
		candidate := filepath.Join(abs, ".git")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}

		parent := filepath.Dir(abs)
		if parent == abs {
			return "", fmt.Errorf(".git directory not found")
		}
		abs = parent
	}
}
