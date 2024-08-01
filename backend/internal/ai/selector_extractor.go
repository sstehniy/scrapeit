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

type EvaluationRequest struct {
	HTML                        string                              `json:"HTML"`
	FieldsToExtractSelectorsFor []models.FieldToExtractSelectorsFor `json:"FieldsToExtractSelectorsFor"`
	ExtractedSelectors          ExtractSelectorsResponse            `json:"ExtractedSelectors"`
}

type EvaluationIssue struct {
	Key    string `json:"key"`
	Remark string `json:"remark"`
}

type EvaluationResponse struct {
	Success bool              `json:"Success"`
	Issues  []EvaluationIssue `json:"Issues"`
}

type ExtractRequestWithEvaluationAttempt struct {
	Output           ExtractSelectorsResponse `json:"Output"`
	EvaluationResult EvaluationResponse       `json:"EvaluationResult"`
}

type ExtractRequestWithEvaluation struct {
	HTML                       string                                `json:"HTML"`
	FieldToExtractSelectorsFor []models.FieldToExtractSelectorsFor   `json:"FieldsToExtractSelectorsFor"`
	PreviousAttempts           []ExtractRequestWithEvaluationAttempt `json:"PreviousAttempts"`
}

func getSystemPromptFromFile() string {
	file, err := os.ReadFile("./system_prompt.txt")
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return string(file)
}

func getEvalSystemPrompt() string {
	file, err := os.ReadFile("./evaluation_prompt.txt")
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return string(file)
}

func getSystemPromptWithEval() string {
	file, err := os.ReadFile("./system_prompt_with_eval.txt")
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return string(file)
}

// func logToFile(prefix string, content string) error {
// 	fileName := fmt.Sprintf("%s.log", prefix)
// 	return os.WriteFile(fileName, []byte(content), 0644)
// }

func makeOpenAICall(dialogue []openai.ChatCompletionMessage, output interface{}, prefix string) (int, int, error) {
	ctx := context.Background()

	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	tokenCount, err := countTokens(dialogue, "gpt-4o")

	if err != nil {
		fmt.Printf("Error counting tokens: %v\n", err)
	} else {
		fmt.Printf("Input Token count for request: %d\n", tokenCount)
	}

	// Log the request
	// requestJSON, _ := json.MarshalIndent(dialogue, "", "  ")
	// if err := logToFile(prefix+"_request", string(requestJSON)); err != nil {
	// 	fmt.Printf("Error logging request: %v\n", err)
	// }

	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    "gpt-4o",
		Messages: dialogue,
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		},
		MaxTokens: 4096,
	})
	if err != nil {
		fmt.Println(err)
		return 0, 0, err
	}

	// Log the response
	// responseJSON, _ := json.MarshalIndent(resp.Choices[0].Message.Content, "", "  ")
	// if err := logToFile(prefix+"_response", string(responseJSON)); err != nil {
	// 	fmt.Printf("Error logging response: %v\n", err)
	// }

	outputTokens, err := countTokens([]openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleAssistant, Content: resp.Choices[0].Message.Content}}, "gpt-4o")

	if err != nil {
		fmt.Printf("Error counting response tokens: %v\n", err)
	} else {
		fmt.Printf("Output Token count for response: %d\n", outputTokens)
	}

	dataToValidate := []byte(resp.Choices[0].Message.Content)
	err = json.Unmarshal(dataToValidate, output)
	if err != nil {
		fmt.Println(err)
		return 0, 0, err
	}

	return tokenCount, outputTokens, nil
}

func ExtractSelectors(html string, fieldsToExtract []models.FieldToExtractSelectorsFor) (ExtractSelectorsResponse, float32, error) {
	var totalInputTokens, totalOutputTokens int

	fieldsToExtractJsonString, err := json.Marshal(fieldsToExtract)
	if err != nil {
		fmt.Println(err)
		return ExtractSelectorsResponse{}, 0, err
	}

	fieldsToExtractString := string(fieldsToExtractJsonString)
	fmt.Printf("Fields to extract: %v\n", fieldsToExtractString)

	fmt.Printf(`{HTML: %v, FieldsToExtractSelectorsFor:
    %v}`, html, fieldsToExtractString)

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

	var initialResponse ExtractSelectorsResponse

	inputTokens, outputTokens, err := makeOpenAICall(dialogue, &initialResponse, "extract")

	if err != nil {
		return ExtractSelectorsResponse{}, 0, err
	}

	totalInputTokens += inputTokens
	totalOutputTokens += outputTokens

	evalInput := EvaluationRequest{
		HTML:                        html,
		FieldsToExtractSelectorsFor: fieldsToExtract,
		ExtractedSelectors:          initialResponse,
	}

	evalInputBytes, _ := json.Marshal(evalInput)

	evalInputString := string(evalInputBytes)

	evalDialogue := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: getEvalSystemPrompt(),
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: evalInputString,
		},
	}

	var evalResponse EvaluationResponse

	evalInputTokens, evalOutputTokens, err := makeOpenAICall(evalDialogue, &evalResponse, "eval")

	if err != nil {
		return ExtractSelectorsResponse{}, 0, err
	}

	totalInputTokens += evalInputTokens
	totalOutputTokens += evalOutputTokens

	if !evalResponse.Success {
		totalAttempts := 1
		const MAX_ATTEMPTS = 2

		advancedInput := ExtractRequestWithEvaluation{
			HTML:                       html,
			FieldToExtractSelectorsFor: fieldsToExtract,
			PreviousAttempts: []ExtractRequestWithEvaluationAttempt{
				{
					Output:           initialResponse,
					EvaluationResult: evalResponse,
				},
			},
		}

		for totalAttempts <= MAX_ATTEMPTS {
			fmt.Println("Total attempts: ", totalAttempts)
			advancedInputBytes, _ := json.Marshal(advancedInput)

			var advancedExtractResponse ExtractSelectorsResponse

			advancedInputString := string(advancedInputBytes)

			advancedDialogue := []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: getSystemPromptWithEval(),
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: advancedInputString,
				},
			}

			advancedInputTokens, advancedOutputTokens, err := makeOpenAICall(advancedDialogue, &advancedExtractResponse, fmt.Sprint(totalAttempts)+"advanced")
			if err != nil {
				return initialResponse, 0, err
			}
			totalInputTokens += advancedInputTokens
			totalOutputTokens += advancedOutputTokens

			evalInput := EvaluationRequest{
				HTML:                        html,
				FieldsToExtractSelectorsFor: fieldsToExtract,
				ExtractedSelectors:          advancedExtractResponse,
			}

			evalInputBytes, _ := json.Marshal(evalInput)

			evalInputString := string(evalInputBytes)

			evalDialogue := []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: getEvalSystemPrompt(),
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: evalInputString,
				},
			}

			var evalResponse EvaluationResponse

			evalInputTokens, evalOutputTokens, err := makeOpenAICall(evalDialogue, &evalResponse, fmt.Sprint(totalAttempts)+"eval")

			if err != nil {
				return ExtractSelectorsResponse{}, 0, err
			}

			totalInputTokens += evalInputTokens
			totalOutputTokens += evalOutputTokens

			if evalResponse.Success {
				initialResponse = advancedExtractResponse
				break
			}

			advancedInput.PreviousAttempts = []ExtractRequestWithEvaluationAttempt{
				{Output: advancedExtractResponse,
					EvaluationResult: evalResponse,
				}}
			totalAttempts += 1
		}
	}

	return initialResponse, float32(totalInputTokens)*0.15/1000000 + float32(totalOutputTokens)*0.6/1000000, nil
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
