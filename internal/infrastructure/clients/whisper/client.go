package whisper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/anderson-reinaldo/go-explicAI/internal/infrastructure/clients"
)

const basePath = "/v1/audio/transcriptions"

type Client struct {
	HttpClient  *clients.BaseHTTP
	ApiKey      string
	ServiceName string
	Model       string
}

type Response struct {
	Text string `json:"text,omitempty"`
}

func NewClient(serviceName, URL, apiKey, model string, timeout int64) *Client {
	return &Client{
		ServiceName: serviceName,
		HttpClient:  clients.NewHttpClient(URL, timeout),
		ApiKey:      apiKey,
		Model:       model,
	}
}

func (c *Client) Transcribe(ctx context.Context, audio []byte) (*string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", "audio.mp3")
	if err != nil {
		return nil, fmt.Errorf("error on whisper request: error=%s", err.Error())
	}

	_, err = io.Copy(part, bytes.NewReader(audio))
	if err != nil {
		return nil, fmt.Errorf("error on whisper request: error=%s", err.Error())
	}

	_ = writer.WriteField("model", c.Model)
	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("error on whisper request: error=%s", err.Error())
	}

	req := c.HttpClient.Client.
		SetHeader("Authorization", "Bearer "+c.ApiKey).
		SetHeader("Content-Type", writer.FormDataContentType()).
		SetBody(body.Bytes())

	res, err := req.Post(basePath)

	if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("error on whisper request:  response=%s | status=%s", res.Body(), res.Status())
	}

	if err != nil {
		return nil, fmt.Errorf("error on whisper request: error=%s", err.Error())
	}

	var response Response
	if err = json.Unmarshal(res.Body(), &response); err != nil {
		return nil, fmt.Errorf("error on whisper request: error=%s", err.Error())
	}

	if response.Text == "" {
		return nil, fmt.Errorf("error on whisper request: error=empty response")
	}

	return &response.Text, nil
}
