package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"scrapeit/internal/models"

	openai "github.com/sashabaranov/go-openai"
)

type ExtractSelectorsResponse struct {
	Fields []models.FieldSelectorsResponse `json:"fields"`
}

func ExtractSelectors(html string, fieldsToExtract []string) (ExtractSelectorsResponse, error) {

	ctx := context.Background()

	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	systemPrompt := `Your task is to analyze given HTML and extract unique css selectors for fields i want to extract from it.
	If for example i request fields, price_unit and price_value but the selector for the price_unit is the same as the selector for the price_value, you should return the same selector for both fields and create a Regex to extract the corresponding part from the text(e.g. if price_unit and price_value both are texts "100.00 $", price_unit has to have a regex for extracting "$" and price_value should have a regex to extract "100.00"). If you can't find a unique selector for a field, you should return an empty string for the selector. If the value i am looking for is not in text content but set in some attribute, you should include additional field "attributeToGet" with the name of the attribute i should get the value from. If the value is in the text content, you should not include the "attributeToGet" field.
	 Example prompt:
	{
	HTML: <div class="some-class"><span class="title_text">Some title</span><span class="price_text">633 $</span><image class="image" src="..." /></div>
	FieldsToExtractSelectorsFor: ["title", "price", "image"]
	}

	your response should always be in the following json format:
	{"fields": [{"field": "title", "selector": ".some-class .title_text"}, {"field": "price", "selector": ".some-class .price_text"}, ...]}
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

	fmt.Printf("Extracted selectors: %v\n", response)

	return response, nil

}
