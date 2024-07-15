package ai

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/liushuangls/go-anthropic"
	"github.com/sashabaranov/go-openai"
)

func ChatCompletion(prompt string) (string, error) {
	ctx := context.Background()

	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	systemPrompt := "Please provide a minimal and sensible answer to the questions"

	dialogue := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		}, {
			Role:    openai.ChatMessageRoleUser,
			Content: prompt,
		},
	}

	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:     openai.GPT3Dot5Turbo0125,
		Messages:  dialogue,
		MaxTokens: 2048,
	})

	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil

}

func ChatCompletionClaude(prompt string) (string, error) {
	ctx := context.Background()

	client := anthropic.NewClient(os.Getenv("ANTHROPIC_API_KEY"))

	systemPrompt := "Please provide a minimal and sensible answer to the questions"

	dialogue := []anthropic.Message{
		anthropic.NewUserTextMessage(prompt),
	}

	resp, err := client.CreateMessages(ctx, anthropic.MessagesRequest{
		Model:     "claude-3-5-sonnet-20240620",
		Messages:  dialogue,
		System:    systemPrompt,
		MaxTokens: 4096,
	})

	if err != nil {
		var e *anthropic.APIError
		if errors.As(err, &e) {
			fmt.Printf("Messages error, type: %s, message: %s", e.Type, e.Message)
		} else {
			fmt.Printf("Messages error: %v\n", err)
		}
		return "", err
	}

	return resp.Content[0].Text, nil

}
