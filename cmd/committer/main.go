package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

var suggestFlag = &cli.BoolFlag{
	Name:  "suggest",
	Usage: "Print suggested corrections when validation fails",
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
		Flags: []cli.Flag{formatFlag, suggestFlag},
		// Default action: hook mode — committer <path-to-commit-msg-file>
		Action: func(c *cli.Context) error {
			if c.NArg() < 1 {
				return cli.Exit("unexpected error: no arguments passed to this script", 1)
			}
			return runHook(c.Args().First(), cfg, c.String("format"), c.Bool("suggest"))
		},
		Commands: []*cli.Command{
			{
				Name:      "lint",
				Usage:     "Validate a commit message string",
				ArgsUsage: "<message>",
				Flags:     []cli.Flag{formatFlag, suggestFlag},
				Action: func(c *cli.Context) error {
					if c.NArg() < 1 {
						return cli.Exit("lint requires a commit message argument", 1)
					}
					msg := committer.Message(c.Args().First())
					return runValidate(msg, cfg, c.String("format"), c.Bool("suggest"))
				},
			},
			{
				Name:  "install",
				Usage: "Install the commit-msg hook into the nearest .git directory",
				Action: func(c *cli.Context) error {
					return runInstall()
				},
			},
			{
				Name:  "init",
				Usage: "Create .committer.yaml config, install git hook, and create .githooks/commit-msg",
				Action: func(c *cli.Context) error {
					return runInit()
				},
			},
			{
				Name:  "check",
				Usage: "Validate recent commit messages in the current git repository",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:  "count",
						Value: 10,
						Usage: "Number of recent commits to validate",
					},
					formatFlag,
				},
				Action: func(c *cli.Context) error {
					return runCheck(c.Int("count"), cfg, c.String("format"))
				},
			},
			{
				Name:  "config",
				Usage: "Manage committer-go configuration",
				Subcommands: []*cli.Command{
					{
						Name:  "validate",
						Usage: "Validate .committer.yaml in the current directory",
						Action: func(c *cli.Context) error {
							return runConfigValidate()
						},
					},
					{
						Name:  "schema",
						Usage: "Print JSON Schema for .committer.yaml to stdout",
						Action: func(c *cli.Context) error {
							return runConfigSchema()
						},
					},
				},
			},
			{
				Name:  "changelog",
				Usage: "Generate a CHANGELOG from conventional commit history",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "from",
						Usage: "Start ref (tag or commit hash); defaults to last git tag",
					},
					&cli.StringFlag{
						Name:  "to",
						Value: "HEAD",
						Usage: "End ref (tag or commit hash)",
					},
					&cli.StringFlag{
						Name:  "output",
						Usage: "Write output to file instead of stdout",
					},
					&cli.StringFlag{
						Name:  "version",
						Usage: "Version label for the changelog section header",
					},
					&cli.BoolFlag{
						Name:  "strict",
						Usage: "Fail if any commit in range is not a valid Conventional Commit",
					},
				},
				Action: func(c *cli.Context) error {
					return runChangelog(
						c.String("from"),
						c.String("to"),
						c.String("output"),
						c.String("version"),
						c.Bool("strict"),
					)
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		os.Exit(1)
	}
}

// runHook is the original hook-mode logic.
func runHook(path string, cfg committer.Config, format string, suggest bool) error {
	commitMsg, err := committer.Read(path)
	if err != nil {
		committer.Print(commitMsg, err.Error())
		return cli.Exit("", 1)
	}
	return runValidate(commitMsg, cfg, format, suggest)
}

// runValidate validates msg and prints results in the requested format.
func runValidate(msg committer.Message, cfg committer.Config, format string, suggest bool) error {
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
		if suggest {
			suggestions := committer.Suggest(msg, cfg)
			fmt.Print(committer.FormatSuggestions(suggestions))
		}
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

const committerYAMLTemplate = `# committer-go configuration
# https://github.com/TangoEnSkai/committer-go

length:
  min: 10   # minimum header length in characters
  max: 60   # maximum header length in characters

types:
  extra: [] # add project-specific types here, e.g. [wip, release]
`

const githookContent = `#!/bin/sh
committer "$1"
`

// runInit creates .committer.yaml, installs the git hook, and creates .githooks/commit-msg.
func runInit() error {
	// 1. Create .committer.yaml
	const configFile = ".committer.yaml"
	if _, err := os.Stat(configFile); err == nil {
		fmt.Printf("%s already exists, skipping\n", configFile)
	} else {
		if err := os.WriteFile(configFile, []byte(committerYAMLTemplate), 0o644); err != nil {
			return cli.Exit(fmt.Sprintf("could not write %s: %v", configFile, err), 1)
		}
		fmt.Printf("created %s\n", configFile)
	}

	// 2. Install git hook
	if err := runInstall(); err != nil {
		return err
	}

	// 3. Create .githooks/commit-msg for team sharing
	if err := os.MkdirAll(".githooks", 0o755); err != nil {
		return cli.Exit(fmt.Sprintf("could not create .githooks directory: %v", err), 1)
	}
	githookPath := filepath.Join(".githooks", "commit-msg")
	if err := os.WriteFile(githookPath, []byte(githookContent), 0o755); err != nil {
		return cli.Exit(fmt.Sprintf("could not write %s: %v", githookPath, err), 1)
	}
	fmt.Printf("created %s for team sharing\n", githookPath)
	fmt.Println()
	fmt.Println("To share the hook with your team, commit .githooks/ and ask teammates to run:")
	fmt.Println("  git config core.hooksPath .githooks")
	return nil
}

// runConfigValidate validates .committer.yaml in the current working directory.
func runConfigValidate() error {
	cfg, err := committer.LoadConfig()
	if err != nil {
		return cli.Exit(fmt.Sprintf(".committer.yaml is invalid: %v", err), 1)
	}
	// Basic sanity checks on the loaded config.
	if cfg.Length.Min <= 0 {
		return cli.Exit("config error: length.min must be > 0", 1)
	}
	if cfg.Length.Max <= 0 {
		return cli.Exit("config error: length.max must be > 0", 1)
	}
	if cfg.Length.Min > cfg.Length.Max {
		return cli.Exit(fmt.Sprintf("config error: length.min (%d) must be <= length.max (%d)", cfg.Length.Min, cfg.Length.Max), 1)
	}
	fmt.Println(".committer.yaml is valid")
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(cfg)
	return nil
}

// committerYAMLSchema is the JSON Schema for .committer.yaml.
const committerYAMLSchema = `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "committer-go configuration",
  "description": "Configuration file for committer-go commit message validator",
  "type": "object",
  "additionalProperties": false,
  "properties": {
    "length": {
      "type": "object",
      "description": "Controls header length validation",
      "additionalProperties": false,
      "properties": {
        "min": {
          "type": "integer",
          "description": "Minimum header length in characters (default: 10)",
          "minimum": 1
        },
        "max": {
          "type": "integer",
          "description": "Maximum header length in characters (default: 60)",
          "minimum": 1
        }
      }
    },
    "types": {
      "type": "object",
      "description": "Extends the built-in commit type list",
      "additionalProperties": false,
      "properties": {
        "extra": {
          "type": "array",
          "description": "Additional project-specific commit types",
          "items": { "type": "string" }
        }
      }
    },
    "body": {
      "type": "object",
      "description": "Controls body/footer validation",
      "additionalProperties": false,
      "properties": {
        "max_line_length": {
          "type": "integer",
          "description": "Maximum body line length (default: 72, 0 = disabled)",
          "minimum": 0
        }
      }
    },
    "ai": {
      "type": "object",
      "description": "Controls optional Claude AI integration",
      "additionalProperties": false,
      "properties": {
        "enabled": {
          "type": "boolean",
          "description": "Enable AI-assisted commit message suggestions (default: false)"
        },
        "provider": {
          "type": "string",
          "description": "AI provider to use (default: anthropic)",
          "enum": ["anthropic"]
        },
        "model": {
          "type": "string",
          "description": "Model ID to use (default: claude-haiku-4-5-20251001)"
        },
        "api_key_env": {
          "type": "string",
          "description": "Environment variable holding the API key (default: ANTHROPIC_API_KEY)"
        },
        "auto_fix_on_failure": {
          "type": "boolean",
          "description": "Automatically suggest a fix when validation fails (default: false)"
        },
        "max_diff_chars": {
          "type": "integer",
          "description": "Maximum characters of diff to send to the AI (default: 4000)",
          "minimum": 0
        }
      }
    }
  }
}
`

// runConfigSchema prints the JSON Schema for .committer.yaml to stdout.
func runConfigSchema() error {
	fmt.Print(committerYAMLSchema)
	return nil
}

// checkResult holds one commit's validation result for batch mode.
type checkResult struct {
	Subject string `json:"subject"`
	Valid   bool   `json:"valid"`
	Error   string `json:"error,omitempty"`
}

// checkSummary is the JSON output of runCheck.
type checkSummary struct {
	Total   int           `json:"total"`
	Passed  int           `json:"passed"`
	Failed  int           `json:"failed"`
	Results []checkResult `json:"results"`
}

// runCheck validates the last count commits in the current git repo.
func runCheck(count int, cfg committer.Config, format string) error {
	out, err := exec.Command("git", "log", fmt.Sprintf("-%d", count), "--format=%s").Output()
	if err != nil {
		return cli.Exit(fmt.Sprintf("git log failed: %v", err), 1)
	}

	lines := strings.Split(strings.TrimRight(string(out), "\n"), "\n")
	if len(lines) == 1 && lines[0] == "" {
		lines = nil
	}

	var results []checkResult
	passed, failed := 0, 0
	for _, subject := range lines {
		msg := committer.Message(subject)
		errMsg, ok := committer.ValidateMessage(msg, cfg)
		r := checkResult{Subject: subject, Valid: ok}
		if !ok {
			r.Error = errMsg
			failed++
		} else {
			passed++
		}
		results = append(results, r)
	}

	if format == "json" {
		summary := checkSummary{
			Total:   len(results),
			Passed:  passed,
			Failed:  failed,
			Results: results,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(summary); err != nil {
			return cli.Exit(fmt.Sprintf("json encode error: %v", err), 1)
		}
		if failed > 0 {
			return cli.Exit("", 1)
		}
		return nil
	}

	// text mode
	fmt.Printf("Checked %d commits: %d passed, %d failed\n", len(results), passed, failed)
	if failed > 0 {
		fmt.Println("\nFailed commits:")
		for _, r := range results {
			if !r.Valid {
				fmt.Printf("  ✗ %s\n    %s\n", r.Subject, r.Error)
			}
		}
		return cli.Exit("", 1)
	}
	return nil
}


// runChangelog generates a CHANGELOG from conventional commit history.
func runChangelog(from, to, output, version string, strict bool) error {
	entries, skipped, err := committer.GitLog(from, to, strict)
	if err != nil {
		return cli.Exit(fmt.Sprintf("changelog: %v", err), 1)
	}

	if len(skipped) > 0 {
		fmt.Fprintf(os.Stderr, "warning: skipped %d non-conventional commits\n", len(skipped))
	}

	if len(entries) == 0 {
		fmt.Fprintln(os.Stderr, "no conventional commits found in range")
		return nil
	}

	sections := committer.BuildChangelog(entries, true)
	md := committer.FormatChangelog(sections, version)

	if output != "" {
		if err := os.WriteFile(output, []byte(md), 0o644); err != nil {
			return cli.Exit(fmt.Sprintf("could not write %s: %v", output, err), 1)
		}
		fmt.Printf("wrote changelog to %s\n", output)
		return nil
	}

	fmt.Print(md)
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
