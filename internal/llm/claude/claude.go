package claude

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/saurabh0719/kiwi/internal/llm/core"
	"github.com/saurabh0719/kiwi/internal/tools"
	"github.com/saurabh0719/kiwi/internal/util"
)

// Adapter implements the Adapter interface for Claude
type Adapter struct {
	client anthropic.Client
	model  string
	tools  *tools.Registry
}

// New creates a new Claude adapter
func New(model, apiKey string, tools *tools.Registry) (*Adapter, error) {
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("Anthropic API key not found")
		}
	}

	client := anthropic.NewClient(
		option.WithAPIKey(apiKey),
	)

	return &Adapter{
		client: client,
		model:  model,
		tools:  tools,
	}, nil
}

// Chat sends a message to Claude and returns the response
func (a *Adapter) Chat(ctx context.Context, messages []core.Message) (string, error) {
	response, _, err := a.ChatWithMetrics(ctx, messages)
	return response, err
}

// prepareTools converts the available tools to Anthropic's tool format
func (a *Adapter) prepareTools() []anthropic.ToolParam {
	if a.tools == nil {
		return nil
	}

	var tools []anthropic.ToolParam
	toolsList := a.tools.List()

	for _, tool := range toolsList {
		// Create a JSON schema for the function parameters
		params := tool.Parameters()

		// Create a proper ToolInputSchemaParam
		var properties = make(map[string]interface{})
		var requiredProps []string

		for name, param := range params {
			properties[name] = map[string]interface{}{
				"type":        param.Type,
				"description": param.Description,
			}

			if param.Required {
				requiredProps = append(requiredProps, name)
			}
		}

		// Create a tool definition with correct schema structure
		toolSchema := anthropic.ToolInputSchemaParam{
			Properties: properties,
		}

		// Add required fields to extraFields
		if len(requiredProps) > 0 {
			toolSchema.ExtraFields = map[string]interface{}{
				"required": requiredProps,
			}
		}

		tools = append(tools, anthropic.ToolParam{
			Name:        tool.Name(),
			Description: anthropic.String(tool.Description()),
			InputSchema: toolSchema,
		})
	}

	return tools
}

// convertMessages converts core.Message slice to Anthropic's Message format
func (a *Adapter) convertMessages(messages []core.Message) []anthropic.MessageParam {
	var claudeMessages []anthropic.MessageParam

	for i, msg := range messages {
		// Skip the first message if it's a system message (handled separately)
		if i == 0 && msg.Role == "system" {
			continue
		}

		var content []anthropic.ContentBlockParamUnion
		content = append(content, anthropic.ContentBlockParamOfRequestTextBlock(msg.Content))

		// Map the message role
		switch msg.Role {
		case "user":
			claudeMessages = append(claudeMessages, anthropic.MessageParam{
				Role:    anthropic.MessageParamRoleUser,
				Content: content,
			})
		case "assistant":
			claudeMessages = append(claudeMessages, anthropic.MessageParam{
				Role:    anthropic.MessageParamRoleAssistant,
				Content: content,
			})
		}
	}

	return claudeMessages
}

// getSystemPrompt extracts the system prompt from the messages
func (a *Adapter) getSystemPrompt(messages []core.Message) []anthropic.TextBlockParam {
	// Get system prompt content
	var systemText string

	// Check if the first message is a system message
	if len(messages) > 0 && messages[0].Role == "system" {
		systemText = messages[0].Content
	} else {
		// Use the default system prompt
		systemText = core.DefaultSystemPrompt
		if a.tools != nil {
			systemText += "\n\n" + a.tools.GetToolsDescription()
		}
	}

	// Return as TextBlockParam slice
	return []anthropic.TextBlockParam{
		{Text: systemText},
	}
}

// ChatWithMetrics sends a message to Claude and returns the response with metrics
func (a *Adapter) ChatWithMetrics(ctx context.Context, messages []core.Message) (string, *core.ResponseMetrics, error) {
	startTime := time.Now()

	// Extract the system prompt as TextBlockParam slice
	systemBlocks := a.getSystemPrompt(messages)

	// Convert messages to Claude format
	claudeMessages := a.convertMessages(messages)

	// Prepare tools if available
	var toolsUnion []anthropic.ToolUnionParam
	if a.tools != nil && len(a.tools.List()) > 0 {
		tools := a.prepareTools()
		if len(tools) > 0 {
			for _, tool := range tools {
				toolsUnion = append(toolsUnion, anthropic.ToolUnionParam{
					OfTool: &tool,
				})
			}
		}
	}

	// Create the message request
	req := anthropic.MessageNewParams{
		Model:     a.model,
		Messages:  claudeMessages,
		System:    systemBlocks,
		MaxTokens: 2000,
	}

	// Add tools if available
	if len(toolsUnion) > 0 {
		req.Tools = toolsUnion
	}

	// Maximum number of tool call iterations to prevent infinite loops
	maxCalls := 5
	callCount := 0

	var finalResponse string

	// Process the conversation with potential tool calls
	for callCount < maxCalls {
		callCount++

		resp, err := a.client.Messages.New(ctx, req)
		if err != nil {
			return "", nil, fmt.Errorf("failed to create message: %w", err)
		}

		// Check if there are tool calls in the response
		hasToolCalls := false
		var toolResults []anthropic.ContentBlockParamUnion

		// Process each block in the response
		for _, block := range resp.Content {
			// Get the variant of the content block
			switch blockVariant := block.AsAny().(type) {
			case anthropic.TextBlock:
				// Collect text from text blocks
				finalResponse += blockVariant.Text
			case anthropic.ToolUseBlock:
				// Found a tool use block
				hasToolCalls = true

				// Get the tool to execute
				toolName := blockVariant.Name
				toolID := blockVariant.ID

				tool, exists := a.tools.Get(toolName)
				if !exists {
					toolResults = append(toolResults,
						anthropic.NewToolResultBlock(toolID,
							fmt.Sprintf("Error: Tool %s not found", toolName),
							true))
					continue
				}

				// Parse the tool input
				var args map[string]interface{}
				toolInputStr := blockVariant.JSON.Input.Raw()

				if err := json.Unmarshal([]byte(toolInputStr), &args); err != nil {
					toolResults = append(toolResults,
						anthropic.NewToolResultBlock(toolID,
							fmt.Sprintf("Error parsing arguments: %v", err),
							true))
					continue
				}

				// Execute the tool with visual feedback
				output, err := util.ExecuteToolWithFeedback(ctx, tool, args)
				if err != nil {
					toolResults = append(toolResults,
						anthropic.NewToolResultBlock(toolID,
							fmt.Sprintf("Error executing tool: %v", err),
							true))
					continue
				}

				// Add the successful result
				toolResults = append(toolResults,
					anthropic.NewToolResultBlock(toolID, output, false))
			}
		}

		// If there are no tool calls, we're done
		if !hasToolCalls {
			break
		}

		// Create a new user message with the tool results
		userMessage := anthropic.MessageParam{
			Role:    anthropic.MessageParamRoleUser,
			Content: toolResults,
		}

		// Add the assistant's message and tool results to continue the conversation
		req.Messages = append(req.Messages, resp.ToParam(), userMessage)

		// Reset final response for the next iteration
		finalResponse = ""
	}

	responseTime := time.Since(startTime)

	// Claude API doesn't provide token usage, estimate based on length
	promptLen := 0
	for _, block := range systemBlocks {
		promptLen += len(block.Text)
	}

	for _, msg := range claudeMessages {
		for _, block := range msg.Content {
			if block.OfRequestTextBlock != nil {
				promptLen += len(block.OfRequestTextBlock.Text)
			}
		}
	}

	// This is a simplified approach estimating token usage (roughly 4 chars per token)
	estimatedPromptTokens := promptLen / 4
	estimatedCompletionTokens := len(finalResponse) / 4

	metrics := &core.ResponseMetrics{
		PromptTokens:     estimatedPromptTokens,
		CompletionTokens: estimatedCompletionTokens,
		TotalTokens:      estimatedPromptTokens + estimatedCompletionTokens,
		ResponseTime:     responseTime,
	}

	return finalResponse, metrics, nil
}

// GetModel returns the model name being used
func (a *Adapter) GetModel() string {
	return a.model
}

// GetProvider returns the provider name
func (a *Adapter) GetProvider() string {
	return "claude"
}

// Name returns the name of the adapter
func (a *Adapter) Name() string {
	return "Claude"
}
