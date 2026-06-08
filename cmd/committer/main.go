package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/TangoEnSkai/committer-go/committer"
	"github.com/urfave/cli/v2"
)

const hookContent = `#!/bin/sh
committer "$1"
`

var formatFlag = &cli.StringFlag{
	Name:  "format",
	Value: "text",
	Usage: "Output format: text or json",
}

func main() {
	cfg, err := committer.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not load config: %v\n", err)
		cfg = committer.DefaultConfig()
	}

	app := &cli.App{
		Name:  "committer",
		Usage: "Conventional Commits message validator",
		Flags: []cli.Flag{formatFlag},
		// Default action: hook mode — committer <path-to-commit-msg-file>
		Action: func(c *cli.Context) error {
			if c.NArg() < 1 {
				return cli.Exit("unexpected error: no arguments passed to this script", 1)
			}
			return runHook(c.Args().First(), cfg, c.String("format"))
		},
		Commands: []*cli.Command{
			{
				Name:      "lint",
				Usage:     "Validate a commit message string",
				ArgsUsage: "<message>",
				Flags:     []cli.Flag{formatFlag},
				Action: func(c *cli.Context) error {
					if c.NArg() < 1 {
						return cli.Exit("lint requires a commit message argument", 1)
					}
					msg := committer.Message(c.Args().First())
					return runValidate(msg, cfg, c.String("format"))
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
		os.Exit(1)
	}
}

// runHook is the original hook-mode logic.
func runHook(path string, cfg committer.Config, format string) error {
	commitMsg, err := committer.Read(path)
	if err != nil {
		committer.Print(commitMsg, err.Error())
		return cli.Exit("", 1)
	}
	return runValidate(commitMsg, cfg, format)
}

// runValidate validates msg and prints results in the requested format.
func runValidate(msg committer.Message, cfg committer.Config, format string) error {
	if format == "json" {
		result := committer.ValidateMessageStructured(msg, cfg)
		enc := json.NewEncoder(os.Stdout)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(result); err != nil {
			return cli.Exit(fmt.Sprintf("json encode error: %v", err), 1)
		}
		if !result.Valid {
			return cli.Exit("", 1)
		}
		return nil
	}

	// text mode (default)
	errMsg, ok := committer.ValidateMessage(msg, cfg)
	if !ok {
		committer.Print(msg, errMsg)
		return cli.Exit("", 1)
	}
	fmt.Println("commit message is valid")
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
