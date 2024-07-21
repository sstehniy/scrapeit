package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"scrapeit/internal/models"

	"github.com/pkoukk/tiktoken-go"
	openai "github.com/sashabaranov/go-openai"
)

type ExtractSelectorsResponse struct {
	Fields []models.FieldSelectorsResponse `json:"fields"`
}

func getSystemPromptFromFile() string {
	file, err := os.ReadFile("/app/internal/ai/system_prompt.txt")
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return string(file)
}

func ExtractSelectors(html string, fieldsToExtract []models.FieldToExtractSelectorsFor) (ExtractSelectorsResponse, error) {

	fieldsToExtractJsonString, err := json.Marshal(fieldsToExtract)
	if err != nil {
		fmt.Println(err)
		return ExtractSelectorsResponse{}, err

	}

	fieldsToExtractString := string(fieldsToExtractJsonString)
	fmt.Printf("Fields to extract: %v\n", fieldsToExtractString)

	ctx := context.Background()

	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	dialogue := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: getSystemPromptFromFile(),
		},
		{
			Role: openai.ChatMessageRoleUser,
			Content: fmt.Sprintf(`{HTML: %v, FieldsToExtractSelectorsFor:
				%v}`, html, fieldsToExtractString),
		},
	}
	tokenCount, err := countTokens(dialogue, "gpt-4o-mini")
	if err != nil {
		return ExtractSelectorsResponse{}, fmt.Errorf("error counting tokens: %w", err)
	}
	fmt.Printf("Token count for request: %d\n", tokenCount)
	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    "gpt-4o-mini",
		Messages: dialogue,
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		},
		MaxTokens: 4096,
	})
	if err != nil {
		fmt.Println(err)
		return ExtractSelectorsResponse{}, err
	}
	fmt.Println(resp.Choices[0].Message.Content)

	dataToValidate := []byte(resp.Choices[0].Message.Content)
	var response ExtractSelectorsResponse
	err = json.Unmarshal(dataToValidate, &response)
	if err != nil {
		fmt.Println(err)
		return ExtractSelectorsResponse{}, err
	}

	responseTokens, err := countTokens([]openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleAssistant, Content: resp.Choices[0].Message.Content}}, "gpt-4o-mini")

	if err != nil {
		fmt.Printf("Error counting response tokens: %v\n", err)
	} else {
		fmt.Printf("Token count for response: %d\n", responseTokens)
	}

	fmt.Printf("Input cost: %.2f, Output cost: %.2f", float32(tokenCount)*0.15/1000000, float32(responseTokens)*0.6/1000000)

	return response, nil

}

func countTokens(messages []openai.ChatCompletionMessage, model string) (int, error) {
	tkm, err := tiktoken.EncodingForModel(model)
	if err != nil {
		return 0, err
	}

	var tokenCount int
	for _, message := range messages {
		tokenCount += len(tkm.Encode(message.Content, nil, nil))
		tokenCount += len(tkm.Encode(message.Role, nil, nil))
		// Add 3 tokens for message metadata
		tokenCount += 3
	}

	return tokenCount, nil
}
