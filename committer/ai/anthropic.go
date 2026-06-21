package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/packages/param"
)

// AnthropicProvider implements AIProvider using the Anthropic Claude API.
type AnthropicProvider struct {
	client *anthropic.Client
	model  string
}

// NewAnthropicProvider creates an AnthropicProvider using the API key from the given env var.
func NewAnthropicProvider(apiKeyEnv, model string) (*AnthropicProvider, error) {
	key := os.Getenv(apiKeyEnv)
	if key == "" {
		return nil, fmt.Errorf("%w: %s is not set", ErrProviderUnavailable, apiKeyEnv)
	}
	if model == "" {
		model = "claude-haiku-4-5-20251001"
	}
	client := anthropic.NewClient(option.WithAPIKey(key))
	return &AnthropicProvider{client: &client, model: model}, nil
}

// commitMessageTool returns the Tool Use definition that enforces structured output.
func commitMessageTool(allowedTypes []string) anthropic.ToolUnionParam {
	typeEnum := make([]interface{}, len(allowedTypes))
	for i, t := range allowedTypes {
		typeEnum[i] = t
	}
	tool := anthropic.ToolParam{
		Name:        "commit_message",
		Description: param.NewOpt("Generate a Conventional Commits message"),
		InputSchema: anthropic.ToolInputSchemaParam{
			Properties: map[string]interface{}{
				"type": map[string]interface{}{
					"type":        "string",
					"description": "The commit type",
					"enum":        typeEnum,
				},
				"scope": map[string]interface{}{
					"type":        "string",
					"description": "Optional scope (e.g. auth, api)",
				},
				"breaking": map[string]interface{}{
					"type":        "boolean",
					"description": "Whether this is a breaking change",
				},
				"description": map[string]interface{}{
					"type":        "string",
					"description": "Short description, imperative mood, no period at end",
				},
				"body": map[string]interface{}{
					"type":        "string",
					"description": "Optional longer body (blank line separates from header)",
				},
			},
			Required: []string{"type", "description"},
		},
	}
	return anthropic.ToolUnionParam{OfTool: &tool}
}

func truncate(s string, maxChars int) string {
	if len(s) <= maxChars {
		return s
	}
	return s[:maxChars] + "\n... (truncated)"
}

func (p *AnthropicProvider) buildSystemPrompt(allowedTypes []string, maxHeaderLen int) string {
	lenStr := "72"
	if maxHeaderLen > 0 {
		lenStr = fmt.Sprintf("%d", maxHeaderLen)
	}
	return fmt.Sprintf(`You are a Git commit message assistant. Analyse the provided diff and generate a single Conventional Commits message.
Allowed types: %s
Keep the header under %s characters. Use imperative mood ("add", not "added").
Call the commit_message tool with your answer.`, strings.Join(allowedTypes, ", "), lenStr)
}

// SuggestFromDiff generates a commit message from a git diff.
func (p *AnthropicProvider) SuggestFromDiff(ctx context.Context, diff string, cfg SuggestConfig) (CommitSuggestion, error) {
	if len(cfg.AllowedTypes) == 0 {
		cfg.AllowedTypes = []string{"feat", "fix", "docs", "style", "refactor", "perf", "test", "build", "ci", "chore", "revert", "deps"}
	}
	maxChars := cfg.MaxDiffChars
	if maxChars == 0 {
		maxChars = 4000
	}

	systemPrompt := p.buildSystemPrompt(cfg.AllowedTypes, cfg.MaxHeaderLen)

	msg, err := p.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(p.model),
		MaxTokens: 512,
		System:    []anthropic.TextBlockParam{{Text: systemPrompt}},
		Tools:     []anthropic.ToolUnionParam{commitMessageTool(cfg.AllowedTypes)},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock("Generate a commit message for this diff:\n\n" + truncate(diff, maxChars))),
		},
	})
	if err != nil {
		return CommitSuggestion{}, fmt.Errorf("%w: %v", ErrInvalidResponse, err)
	}

	return extractToolResult(msg)
}

// FixMessage rewrites a commit message that failed validation.
func (p *AnthropicProvider) FixMessage(ctx context.Context, msg string, violations []string, cfg SuggestConfig) (CommitSuggestion, error) {
	if len(cfg.AllowedTypes) == 0 {
		cfg.AllowedTypes = []string{"feat", "fix", "docs", "style", "refactor", "perf", "test", "build", "ci", "chore", "revert", "deps"}
	}

	systemPrompt := fmt.Sprintf(`You are a Git commit message assistant. Fix the provided commit message to comply with Conventional Commits.
Allowed types: %s
Violations to fix:
%s
Call the commit_message tool with the corrected message.`, strings.Join(cfg.AllowedTypes, ", "), strings.Join(violations, "\n"))

	resp, err := p.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(p.model),
		MaxTokens: 512,
		System:    []anthropic.TextBlockParam{{Text: systemPrompt}},
		Tools:     []anthropic.ToolUnionParam{commitMessageTool(cfg.AllowedTypes)},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock("Fix this commit message: " + msg)),
		},
	})
	if err != nil {
		return CommitSuggestion{}, fmt.Errorf("%w: %v", ErrInvalidResponse, err)
	}

	return extractToolResult(resp)
}

// extractToolResult pulls the commit_message tool call from the API response.
func extractToolResult(msg *anthropic.Message) (CommitSuggestion, error) {
	for _, block := range msg.Content {
		if block.Type == "tool_use" && block.Name == "commit_message" {
			var input struct {
				Type        string `json:"type"`
				Scope       string `json:"scope"`
				Breaking    bool   `json:"breaking"`
				Description string `json:"description"`
				Body        string `json:"body"`
			}
			if err := json.Unmarshal(block.Input, &input); err != nil {
				return CommitSuggestion{}, ErrInvalidResponse
			}
			return CommitSuggestion{
				Type:        input.Type,
				Scope:       input.Scope,
				Breaking:    input.Breaking,
				Description: input.Description,
				Body:        input.Body,
			}, nil
		}
	}
	return CommitSuggestion{}, fmt.Errorf("%w: no tool_use block in response", ErrInvalidResponse)
}
