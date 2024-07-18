package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"scrapeit/internal/models"

	"github.com/liushuangls/go-anthropic"
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

func ExtractSelectorsClaude(html string, fieldsToExtract []models.FieldToExtractSelectorsFor) (ExtractSelectorsResponse, error) {
	ctx := context.Background()

	client := anthropic.NewClient(os.Getenv("ANTHROPIC_API_KEY"))

	fieldsToExtractJsonString, err := json.Marshal(fieldsToExtract)
	if err != nil {
		fmt.Println(err)
		return ExtractSelectorsResponse{}, err

	}

	fieldsToExtractString := string(fieldsToExtractJsonString)
	fmt.Printf("Fields to extract: %v\n", fieldsToExtractString)

	responseStart := `{"fields":`

	dialogue := []anthropic.Message{
		anthropic.NewUserTextMessage(fmt.Sprintf(`{HTML: %v, FieldsToExtractSelectorsFor:
			%v}`, html, fieldsToExtractString)),
		anthropic.NewAssistantTextMessage(responseStart),
	}

	resp, err := client.CreateMessages(ctx, anthropic.MessagesRequest{
		Model:     "claude-3-5-sonnet-20240620",
		Messages:  dialogue,
		System:    getSystemPromptFromFile(),
		MaxTokens: 4096,
	})
	if err != nil {
		var e *anthropic.APIError
		if errors.As(err, &e) {
			fmt.Printf("Messages error, type: %s, message: %s", e.Type, e.Message)
		} else {
			fmt.Printf("Messages error: %v\n", err)
		}
		return ExtractSelectorsResponse{}, err
	}
	fmt.Println(responseStart + resp.Content[0].Text)

	dataToValidate := []byte(responseStart + resp.Content[0].Text)
	var response ExtractSelectorsResponse
	err = json.Unmarshal(dataToValidate, &response)
	if err != nil {
		fmt.Println(err)
		return ExtractSelectorsResponse{}, err
	}

	return response, nil
}

func ExtractSelectorsMistral(html string, fieldsToExtract []models.FieldToExtractSelectorsFor) (ExtractSelectorsResponse, error) {
	ctx := context.Background()

	fieldsToExtractJsonString, err := json.Marshal(fieldsToExtract)
	if err != nil {
		fmt.Println(err)
		return ExtractSelectorsResponse{}, err

	}

	fieldsToExtractString := string(fieldsToExtractJsonString)
	fmt.Printf("Fields to extract: %v\n", fieldsToExtractString)

	jsonData, err := json.Marshal(map[string]interface{}{
		"model": "mistral-small-latest",
		"messages": []map[string]interface{}{
			{
				"role":    "system",
				"content": getSystemPromptFromFile(),
			},
			{
				"role": "user",
				"content": fmt.Sprintf(`{HTML: %v, FieldsToExtractSelectorsFor:
					%v}`, html, fieldsToExtractString),
			},
		},
		"response_format": map[string]string{
			"type": "json_object",
		},
	})
	if err != nil {
		return ExtractSelectorsResponse{}, fmt.Errorf("error marshaling request data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.mistral.ai/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return ExtractSelectorsResponse{}, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+"wUtGVjXlkkkqKN7Dz2tjyjWlQSYDa2TV")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return ExtractSelectorsResponse{}, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return ExtractSelectorsResponse{}, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	type MistralResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	var response MistralResponse

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return ExtractSelectorsResponse{}, fmt.Errorf("error decoding response: %w", err)
	}
	fmt.Println(response)

	var extractSelectorsResponse ExtractSelectorsResponse
	err = json.Unmarshal([]byte(response.Choices[0].Message.Content), &extractSelectorsResponse)

	if err != nil {
		return ExtractSelectorsResponse{}, fmt.Errorf("error unmarshaling response: %w", err)
	}

	// resp, err := client.CreateMessages(ctx, anthropic.MessagesRequest{
	// 	Model:     "claude-3-5-sonnet-20240620",
	// 	Messages:  dialogue,
	// 	System:    getSystemPromptFromFile(),
	// 	MaxTokens: 4096,
	// })
	// if err != nil {
	// 	var e *anthropic.APIError
	// 	if errors.As(err, &e) {
	// 		fmt.Printf("Messages error, type: %s, message: %s", e.Type, e.Message)
	// 	} else {
	// 		fmt.Printf("Messages error: %v\n", err)
	// 	}
	// 	return ExtractSelectorsResponse{}, err
	// }
	// fmt.Println(responseStart + resp.Content[0].Text)

	// dataToValidate := []byte(responseStart + resp.Content[0].Text)
	// var response ExtractSelectorsResponse
	// err = json.Unmarshal(dataToValidate, &response)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return ExtractSelectorsResponse{}, err
	// }

	return extractSelectorsResponse, nil
}
