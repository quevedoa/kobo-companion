package llm

import "context"

type GenerateRequest struct {
	Prompt string
}

type GenerateResponse struct {
	Text string
}

type LLM interface {
	Generate(context.Context, GenerateRequest) (*GenerateResponse, error)
}
