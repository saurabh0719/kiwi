package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/saurabh0719/kiwi/internal/config"
	"github.com/saurabh0719/kiwi/internal/input"
	"github.com/saurabh0719/kiwi/internal/llm"
	"github.com/saurabh0719/kiwi/internal/llm/core"
	"github.com/saurabh0719/kiwi/internal/session"
	"github.com/saurabh0719/kiwi/internal/tools"
	"github.com/saurabh0719/kiwi/internal/util"
	"github.com/spf13/cobra"
)

func initChatCmd() {
	chatCmd = &cobra.Command{
		Use:   "chat",
		Short: "Start a new chat session",
		Long:  `Start a new chat session with the LLM. Type 'exit' to end the session.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return startNewChat(cmd, args)
		},
	}
}

func startNewChat(cmd *cobra.Command, args []string) error {
	sessionMgr, err := session.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create session manager: %w", err)
	}

	sessionID := fmt.Sprintf("session_%d", time.Now().Unix())
	sess, err := sessionMgr.CreateSession(sessionID)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	cfg, err := config.Load(cmd)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	toolRegistry := tools.NewRegistry()
	// Register all tools including the websearch tool if API key is available
	tools.RegisterAllTools(toolRegistry, cfg.GetToolsConfig())

	adapter, err := llm.NewAdapter(cfg.LLM.Provider, cfg.LLM.Model, cfg.LLM.APIKey, toolRegistry)
	if err != nil {
		return fmt.Errorf("failed to create LLM adapter: %w", err)
	}

	// session info
	util.InfoColor.Printf("Created new session: %s\n", sessionID)
	fmt.Println("Chat session started. Type 'exit' to end the session. Use Shift+Enter for new lines, Enter to submit")
	util.InfoColor.Printf("Using %s model: %s\n", adapter.GetProvider(), adapter.GetModel())
	util.DividerColor.Println("----------------------------------------")

	for {
		fmt.Println()
		util.UserColor.Print("You: ")

		userInput, err := input.ReadMultiLine("")
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		// Skip empty inputs and request a non-empty message
		if strings.TrimSpace(userInput) == "" {
			// Silently continue the loop to re-prompt the user instead of showing an error
			continue
		}

		if strings.ToLower(strings.TrimSpace(userInput)) == "exit" {
			fmt.Println("Ending session...")
			return nil
		}

		if err := sessionMgr.AddMessage(sess.ID, "user", userInput); err != nil {
			return fmt.Errorf("failed to add user message: %w", err)
		}

		sess, err = sessionMgr.GetSession(sess.ID)
		if err != nil {
			return fmt.Errorf("failed to get updated session: %w", err)
		}

		var messages []llm.Message
		if len(sess.Messages) == 1 {
			messages = append(messages, llm.Message{
				Role: "system",
				Content: `You are Kiwi in chat mode. In this mode, you maintain conversation context and provide thoughtful, helpful responses to user queries over time.

For this chat session:
- Retain context from previous messages
- Provide comprehensive responses when appropriate
- Ask clarifying questions when user requests are ambiguous
- Balance brevity with completeness based on user's engagement style

When interacting with users:
- Be conversational yet efficient
- Show your reasoning when solving complex problems
- Present multiple approaches for complex questions
- Adapt your technical level to match the user's demonstrated expertise

Remember this is an ongoing conversation where context builds over time.`,
			})
		}

		for _, msg := range sess.Messages {
			messages = append(messages, llm.Message{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}

		// Initialize a complete response string to store the entire response
		completeResponse := ""

		// Define a stream handler for handling the streaming response
		streamHandler := func(chunk string) error {
			// Print the chunk without a newline
			fmt.Print(chunk)
			// Append the chunk to the complete response
			completeResponse += chunk
			return nil
		}

		// Start a spinner that will stop once we get the first token
		spinner := util.NewSpinner("Thinking...")
		spinner.Start()

		// Print the assistant prompt
		fmt.Println()

		// Track time for metrics
		startTime := time.Now()
		var metrics *core.ResponseMetrics

		if cfg.UI.Streaming {
			// Use streaming API with the handler
			metrics, err = adapter.ChatStream(context.Background(), messages, func(chunk string) error {
				// Stop the spinner on first token
				if spinner != nil {
					spinner.Stop()
					spinner = nil
					// Print the Kiwi prefix after stopping the spinner
					util.AssistantColor.Print("Kiwi: ")
				}
				return streamHandler(chunk)
			})
		} else {
			// Use non-streaming API for complete response at once
			var response string
			response, metrics, err = adapter.ChatWithMetrics(context.Background(), messages)

			// Stop the spinner when response is received
			if spinner != nil {
				spinner.Stop()
				spinner = nil
			}

			// Print the Kiwi prefix after stopping the spinner
			util.AssistantColor.Print("Kiwi: ")

			// Print the complete response
			fmt.Print(response)
			completeResponse = response
		}

		// If the spinner is still running (no tokens received), stop it
		if spinner != nil {
			spinner.Stop()
		}

		// Print a newline after the response
		fmt.Println()

		if err != nil {
			return fmt.Errorf("failed to get streaming response: %w", err)
		}

		// If metrics is nil (can happen if the stream fails), create empty metrics
		if metrics == nil {
			metrics = &core.ResponseMetrics{
				ResponseTime: time.Since(startTime),
			}
		}

		if cfg.UI.Debug {
			util.StatsColor.Printf("\n[%s] Tokens: %d prompt + %d completion = %d total | Time: %.2fs\n",
				adapter.GetModel(),
				metrics.PromptTokens,
				metrics.CompletionTokens,
				metrics.TotalTokens,
				metrics.ResponseTime.Seconds())
			util.DividerColor.Println("----------------------------------------")
		}

		if err := sessionMgr.AddMessage(sess.ID, "assistant", completeResponse); err != nil {
			return fmt.Errorf("failed to add assistant message: %w", err)
		}
	}
}
