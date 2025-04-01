package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/saurabh0719/kiwi/internal/config"
	"github.com/saurabh0719/kiwi/internal/input"
	"github.com/saurabh0719/kiwi/internal/llm"
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
	tools.RegisterStandardTools(toolRegistry)

	adapter, err := llm.NewAdapter(cfg.LLM.Provider, cfg.LLM.Model, cfg.LLM.APIKey, toolRegistry)
	if err != nil {
		return fmt.Errorf("failed to create LLM adapter: %w", err)
	}

	fmt.Printf("Created new session: %s\n", sessionID)
	fmt.Println("Chat session started. Type 'exit' to end the session.")
	fmt.Printf("Using %s model: %s\n", adapter.GetProvider(), adapter.GetModel())
	fmt.Println("----------------------------------------")

	for {
		fmt.Print("\nYou: ")
		userInput, err := input.ReadMultiLine("")
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
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

		fmt.Print("\nAssistant: ")

		// Start the loading spinner
		spinner := util.NewSpinner("Thinking...")
		spinner.Start()

		response, metrics, err := adapter.ChatWithMetrics(context.Background(), messages)

		// Stop the spinner
		spinner.Stop()

		if err != nil {
			return fmt.Errorf("failed to get response: %w", err)
		}

		fmt.Println(response)

		// Print statistics in blue only when debug mode is enabled
		if cfg.UI.Debug {
			statsColor.Printf("\n[%s] Tokens: %d prompt + %d completion = %d total | Time: %.2fs\n",
				adapter.GetModel(),
				metrics.PromptTokens,
				metrics.CompletionTokens,
				metrics.TotalTokens,
				metrics.ResponseTime.Seconds())
		}

		if err := sessionMgr.AddMessage(sess.ID, "assistant", response); err != nil {
			return fmt.Errorf("failed to add assistant message: %w", err)
		}
	}
}
