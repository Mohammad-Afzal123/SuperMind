package services

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Mohammad-Afzal123/langflow-openai-proxy/types"
	"github.com/Mohammad-Afzal123/langflow-openai-proxy/utils"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func ChatCompletions(c *gin.Context, request types.ChatCompletionsRequest) {
	url := os.Getenv("LANGFLOW_HOST") + "/api/v1/run/" + request.Model + "?stream=" + strconv.FormatBool(request.Stream)
	apiKey := c.GetString("apiKey")

	payload := map[string]interface{}{
		"input_value": request.Messages[len(request.Messages)-1].Content,
		"output_type": "chat",
		"input_type":  "chat",
		"tweaks":      map[string]interface{}{},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Error("Error marshaling payload:", err)
		c.JSON(500, types.ChatCompletionResponse{Error: "Error making request"})
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Error("Error creating request:", err)
		c.JSON(500, types.ChatCompletionResponse{Error: "Error making request"})
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("Error making request:", err)
		c.JSON(500, types.ChatCompletionResponse{Error: "Error making request"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("Error response from server:", resp.StatusCode)
		c.JSON(resp.StatusCode, types.ChatCompletionResponse{Error: "Error making request"})
		return
	}

	if request.Stream {
		type Artifacts struct {
			StreamURL string `json:"stream_url"`
		}

		type Output struct {
			Artifacts Artifacts `json:"artifacts"`
		}

		type RootOutput struct {
			Outputs []Output `json:"outputs"`
		}

		type Response struct {
			Outputs []RootOutput `json:"outputs"`
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Error("Error reading response body:", err)
			c.JSON(500, types.ChatCompletionResponse{Error: "Error making request"})
			return
		}

		var data Response
		err = json.Unmarshal([]byte(body), &data)
		if err != nil {
			log.Error("Error unmarshaling JSON:", err)
			c.JSON(500, types.ChatCompletionResponse{Error: "Error making request"})
			return
		}
		req, err := http.NewRequest("GET", os.Getenv("LANGFLOW_HOST")+data.Outputs[0].Outputs[0].Artifacts.StreamURL, nil)
		if err != nil {
			log.Fatalf("Error creating request: %v", err)
		}

		client := &http.Client{
			Timeout: time.Duration(0),
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Fatalf("Error making request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Fatalf("Unexpected status code: %v", resp.StatusCode)
		}

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			c.SSEvent("message", line)
			c.Writer.Flush()
		}
	} else {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Error("Error reading response body:", err)
			c.JSON(500, types.ChatCompletionResponse{Error: "Error making request"})
			return
		}

		var data types.LangflowResponse
		err = json.Unmarshal([]byte(body), &data)
		if err != nil {
			log.Error("Error unmarshaling JSON:", err)
			c.JSON(500, types.ChatCompletionResponse{Error: "Error making request"})
			return
		}
		var response types.ChatCompletionResponse = types.ChatCompletionResponse{}
		response.Usage.PromptTokens = utils.Tokenize(request.Messages[len(request.Messages)-1].Content)
		response.Usage.CompletionTokens = utils.Tokenize(data.Outputs[0].Outputs[0].Results.Message.Data.Text)
		response.Choices = append(response.Choices, types.Choice{
			Index: 0,
			Message: types.OpenAIMessage{
				Role:    "assistant",
				Content: data.Outputs[0].Outputs[0].Results.Message.Data.Text,
			},
		})
		c.JSON(resp.StatusCode, response)
	}
}
