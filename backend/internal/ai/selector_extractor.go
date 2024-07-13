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

func ExtractSelectors(html string, fieldsToExtract []string) (ExtractSelectorsResponse, error) {

	ctx := context.Background()

	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	systemPrompt := `
		Your task is to analyze given HTML and extract unique, but universal CSS selectors for specified fields. These selectors should work not only for the given HTML but also for similar HTML structures.

Guidelines:
1. If multiple fields share the same selector, use that selector for all of them and provide a regex to extract the specific part for each field.
2. Keep regexes as general as possible. For example, a regex for currency symbols should match both "$" and "€". But try to avoid regex if a simple CSS selector can be used to extract a field for which textContent is needed.
3. If a unique selector can't be found for a field, return an empty string as the selector.
4. If the target value is in an attribute rather than text content, include an "attributeToGet" field specifying the attribute name.
5. Only include "attributeToGet" when the value is in an attribute, not for text content.
6, Avoid using data- attributes selectors as they are brittle and may break if the HTML structure changes. Instead, you can use nth-child or nth-of-type, but try to do your best to avoid using such selectors.

Example input:
{
  "HTML": "<div class=\"some-class\"><span class=\"title_text\">Some title</span><span class=\"price_text\">633 $</span><img class=\"image\" src=\"...\" /></div>",
  "FieldsToExtractSelectorsFor": ["title", "price_value", "price_unit", "image"]
}

Your response should always be in the following JSON format:
{
  "fields": [
    {
      "field": "title",
      "selector": ".some-class .title_text"
    },
    {
      "field": "price_value",
      "selector": ".some-class .price_text",
      "regex": "\\d+(\\.\\d+)?"
    },
    {
      "field": "price_unit",
      "selector": ".some-class .price_text",
      "regex": "[\\$€]"
    },
    {
      "field": "image",
      "selector": ".some-class .image",
      "attributeToGet": "src"
    }
  ]
}

Important notes:
1. Ensure selectors are as specific as necessary but as general as possible to work with similar HTML structures.
2. Prioritize using class selectors over other attributes when possible.
3. If the HTML structure is complex or nested, consider using child (>) or descendant ( ) combinators in selectors as appropriate.
4. Escape special characters in regexes properly for JSON formatting.
5. Test your selectors and regexes to ensure they correctly extract the desired information.
	`
	dialogue := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		},
		{
			Role: openai.ChatMessageRoleUser,
			Content: fmt.Sprintf(`{HTML: %v, FieldsToExtractSelectorsFor: 
				%v}`, html, fieldsToExtract),
		},
	}
	tokenCount, err := countTokens(dialogue, openai.GPT4o)
	if err != nil {
		return ExtractSelectorsResponse{}, fmt.Errorf("error counting tokens: %w", err)
	}
	fmt.Printf("Token count for request: %d\n", tokenCount)
	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    openai.GPT4o,
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

	responseTokens, err := countTokens([]openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleAssistant, Content: resp.Choices[0].Message.Content}}, openai.GPT4)
	if err != nil {
		fmt.Printf("Error counting response tokens: %v\n", err)
	} else {
		fmt.Printf("Token count for response: %d\n", responseTokens)
	}

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
