package chatgpt

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/anderson-reinaldo/go-explicAI/internal/gateway/summarize"
	"github.com/anderson-reinaldo/go-explicAI/internal/infrastructure/clients"
)

const (
	basePath                   = "/v1/chat/completions"
	systemPrompt               = "Você é um sistema que recebe um texto transcrito de um áudio e organiza, separando em parágrafos e corrigindo possíveis erros de concordância."
	resumeUserPrompt           = "Preciso de um objeto com título sugerido, descrição sugerida, resumo breve e médio sobre a seguinte transcrição:"
	fullTextOrganizeUserPrompt = "Retorne apenas o texto normalizado para a seguinte transcrição:"

	functionCallName = "resume"
)

type (
	ChatgptFunctionCallRequest struct {
		Model        string       `json:"model"`
		Messages     []Message    `json:"messages"`
		Functions    []Function   `json:"functions"`
		FunctionCall FunctionCall `json:"function_call,omitempty"`
	}

	ChatgptSimpleRequest struct {
		Model    string    `json:"model"`
		Messages []Message `json:"messages"`
	}

	FunctionCall struct {
		Name string `json:"name"`
	}

	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	Function struct {
		Name       string             `json:"name"`
		Parameters FunctionParameters `json:"parameters,omitempty"`
	}

	FunctionParameters struct {
		Type       string          `json:"type"`
		Properties PropertiesFiels `json:"properties"`
	}

	PropertiesFiels struct {
		Title        PropertieFieldSpec `json:"title"`
		Description  PropertieFieldSpec `json:"description"`
		BriefResume  PropertieFieldSpec `json:"briefResume"`
		MediumResume PropertieFieldSpec `json:"mediumResume"`
	}

	PropertieFieldSpec struct {
		Type        string `json:"type"`
		Description string `json:"description"`
	}

	ChatResumeCompletionsResponse struct {
		Choices []struct {
			Message struct {
				FunctionCall struct {
					Arguments string `json:"arguments"`
				} `json:"function_call"`
			} `json:"message"`
		} `json:"choices"`
	}

	ChatFullTextCompletionResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
)

type Client struct {
	HttpClient  *clients.BaseHTTP
	ApiKey      string
	ServiceName string
	Model       string
}

func NewClient(serviceName, URL, apiKey, model string, timeout int64) *Client {
	return &Client{
		ServiceName: serviceName,
		HttpClient:  clients.NewHttpClient(URL, timeout),
		ApiKey:      apiKey,
		Model:       model,
	}
}

func (c *Client) resume(ctx context.Context, transcription string) (*summarize.ResumeOutput, error) {
	req := c.HttpClient.Client.
		SetHeader("Authorization", "Bearer "+c.ApiKey).
		SetHeader("Content-Type", "aaplication/json").
		SetBody(c.buildResumeRequest(transcription))

	res, err := req.Post(basePath)

	if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("error on chatgpt resume request: response=%s | status=%s", res.Body(), res.Status())
	}

	if err != nil {
		return nil, fmt.Errorf("error on chatgpt resume request: error=%s", err.Error())
	}

	var chatResponse ChatResumeCompletionsResponse
	if err = json.Unmarshal(res.Body(), &chatResponse); err != nil {
		return nil, fmt.Errorf("error on chatgpt resume request: error=%s", err.Error())
	}

	choices := chatResponse.Choices
	if len(choices) <= 0 {
		return nil, fmt.Errorf("error on chatgpt resume request: no choices in response")
	}

	var response summarize.ResumeOutput
	if err = json.Unmarshal([]byte(choices[0].Message.FunctionCall.Arguments), &response); err != nil {
		return nil, fmt.Errorf("error on chatgpt resume request: no choices in response")
	}

	return &response, nil

}

func (c *Client) fullTextOrganize(ctx context.Context, trascription string) (*string, error) {
	req := c.HttpClient.Client.
		SetHeader("Authorization", "Bearer "+c.ApiKey).
		SetHeader("Content-Type", "aaplication/json").
		SetBody(c.buildFullTextOrganizeRequest(trascription))

	res, err := req.Post(basePath)

	if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("error on chatgpt full text organize request: response=%s | status=%s", res.Body(), res.Status())
	}

	if err != nil {
		return nil, fmt.Errorf("error on chatgpt full text organize request: error=%s", err.Error())
	}

	var response ChatFullTextCompletionResponse
	if err = json.Unmarshal(res.Body(), &response); err != nil {
		return nil, fmt.Errorf("error on chatgpt full text organize request: error=%s", err.Error())
	}

	if len(response.Choices) <= 0 || response.Choices[0].Message.Content == "" {
		return nil, fmt.Errorf("error on chatgpt full text organize request: empty response")
	}

	responseText := response.Choices[0].Message.Content

	return &responseText, nil
}

func (c *Client) buildResumeRequest(transcription string) ChatgptFunctionCallRequest {
	fullResumeUserPrompt := resumeUserPrompt + "\n" + transcription

	return ChatgptFunctionCallRequest{
		Model: c.Model,
		Messages: []Message{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: fullResumeUserPrompt,
			},
		},
		Functions: buildFunctionCallRequest(),
		FunctionCall: FunctionCall{
			Name: "resume",
		},
	}
}

func buildFunctionCallRequest() []Function {
	return []Function{
		{
			Name: functionCallName,
			Parameters: FunctionParameters{
				Type: "object",
				Properties: PropertiesFiels{
					Title: PropertieFieldSpec{
						Type:        "string",
						Description: "Título de até 60 caracteres",
					},
					Description: PropertieFieldSpec{
						Type:        "string",
						Description: "Descrição de até 300 caracteres",
					},
					BriefResume: PropertieFieldSpec{
						Type:        "string",
						Description: "Resumo breve de até 5 linhas",
					},
					MediumResume: PropertieFieldSpec{
						Type:        "string",
						Description: "Resumo médio de até 15 linhas",
					},
				},
			},
		},
	}
}

func (c *Client) buildFullTextOrganizeRequest(transcription string) ChatgptSimpleRequest {
	fullTextOrganizePrompt := fullTextOrganizeUserPrompt + "\n" + transcription

	return ChatgptSimpleRequest{
		Model: c.Model,
		Messages: []Message{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: fullTextOrganizePrompt,
			},
		},
	}
}
