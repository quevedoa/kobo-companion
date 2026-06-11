package main

import (
	"kobo-companion/internal/config"
	"kobo-companion/internal/entities"
	jobrepository "kobo-companion/internal/gateway/job-repository"
	"kobo-companion/internal/gateway/llm"
	"kobo-companion/internal/handler"
	"log"
	"net/http"
	"time"
)

var (
	jobs = map[string]*entities.Job{}
)

func main() {
	cfg := config.Load()

	llmGateway := llm.NewOpenAIGateway(llm.OpenAIOptions{
		APIKey: cfg.OpenAIAPIKey,
		Model:  cfg.OpenAIModel,
		Client: &http.Client{Timeout: 30 * time.Second},
	})
	jobRepo := jobrepository.NewInMemoryJobRepo()

	h := handler.New(llmGateway, jobRepo)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/selection", h.HandleSelection)
	mux.HandleFunc("GET /job/{jobID}", h.HandleJob)
	mux.HandleFunc("/latest", h.HandleLatest)

	log.Println("Listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(cfg.ListenAddress, mux))
}
