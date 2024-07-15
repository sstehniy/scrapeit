package handlers

import (
	"net/http"
	"scrapeit/internal/ai"

	"github.com/labstack/echo/v4"
)

type CompletionRequest struct {
	Prompt string `json:"prompt"`
}

func CompletionHandler(c echo.Context) error {

	var body CompletionRequest
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	resp, err := ai.ChatCompletionClaude(body.Prompt)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"response": resp})
}
