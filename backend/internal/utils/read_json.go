package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func ReadJson(path string) ([]byte, error) {
	pwd, _ := os.Getwd()
	jsonData, err := os.Open(pwd + path)
	if err != nil {
		fmt.Println("Error reading JSON groups: ", err)
		return nil, err
	}
	defer jsonData.Close()
	byteValue, _ := io.ReadAll(jsonData)
	return byteValue, nil
}

func WriteJson(path string, data interface{}) error {
	pwd, _ := os.Getwd()
	file, err := os.OpenFile(pwd+path, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Println("Error opening JSON file: ", err)
		return err
	}
	defer file.Close()

	jsonData, err := json.Marshal(data)
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
