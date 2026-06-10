package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type OpenAIOptions struct {
	APIKey string
	Model  string
	Client *http.Client
}

type openAIGateway struct {
	apiKey string
	model  string
	client *http.Client
}

func NewOpenAIGateway(o OpenAIOptions) *openAIGateway {
	return &openAIGateway{
		apiKey: o.APIKey,
		model:  o.Model,
		client: o.Client,
	}
}

func (o *openAIGateway) Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {

	body := map[string]interface{}{
		"model": o.model,
		"input": req.Prompt,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		"POST",
		"https://api.openai.com/v1/responses",
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+o.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	res, err := o.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf(
			"OpenAI responses API returned %s: %s",
			res.Status,
			strings.TrimSpace(string(resBody)),
		)
	}

	openAIRes := struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Output []struct {
			Type    string `json:"type"`
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"output"`
	}{}

	err = json.Unmarshal(resBody, &openAIRes)
	if err != nil {
		return nil, err
	}

	return &GenerateResponse{
		Text: openAIRes.Output[0].Content[0].Text,
	}, nil
}
