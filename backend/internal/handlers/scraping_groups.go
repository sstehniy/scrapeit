package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"scrapeit/internal/models"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func getAllGroups() (*[]models.ScrapeGroup, error) {
	pwd, _ := os.Getwd()
	jsonData, err := os.Open(pwd + "/internal/data/scraping_groups.json")
	if err != nil {
		fmt.Println("Error reading JSON groups: ", err)
		return nil, err
	}

	defer jsonData.Close()

	byteValue, _ := io.ReadAll(jsonData)

	var groups *[]models.ScrapeGroup

	err = json.Unmarshal(byteValue, &groups)

	if err != nil {
		fmt.Println("Error unmarshaling JSON bytes", err)
		return nil, err
	}

	return groups, nil
}

func writeAllGroups(groups *[]models.ScrapeGroup) error {
	pwd, _ := os.Getwd()
	file, err := os.OpenFile(pwd+"/internal/data/scraping_groups.json", os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Println("Error opening JSON file: ", err)
		return err
	}
	defer file.Close()

	jsonData, err := json.Marshal(groups)
	if err != nil {
		fmt.Println("Error marshaling JSON: ", err)
		return err
	}

	_, err = file.Write(jsonData)
	if err != nil {
		fmt.Println("Error writing JSON data: ", err)
		return err
	}

	return nil
}

func GetScrapingGroups(c echo.Context) error {
	allGroups, _ := getAllGroups()

	return c.JSON(http.StatusOK, allGroups)
}

func CreateScrapingGroup(c echo.Context) error {
	body := map[string]interface{}{}
	if err := c.Bind(&body); err != nil {
		return err
	}

	var newGroup *models.ScrapeGroup

	if name, ok := body["name"].(string); ok {
		allGroups, _ := getAllGroups()
		newGroup = &models.ScrapeGroup{
			ID:        uuid.New().String(),
			Name:      name,
			URL:       "",
			Fields:    []models.Field{},
			Endpoints: []models.Endpoint{},
		}
		*allGroups = append(*allGroups, *newGroup)
		err := writeAllGroups(allGroups)
		if err != nil {
			fmt.Println("Error while writing to all groups", err)
			return err
		}

	} else {
		fmt.Println("name is not a string")
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid name type in body",
		})
	}

	return c.JSON(http.StatusOK, newGroup)
}
