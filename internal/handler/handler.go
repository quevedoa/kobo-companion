package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"kobo-companion/internal/entities"
	jobrepository "kobo-companion/internal/gateway/job-repository"
	"kobo-companion/internal/gateway/llm"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type Handler struct {
	LLMGateway llm.LLM
	JobRepo    jobrepository.JobRepository
}

func New(llmGateway llm.LLM, jobRepo jobrepository.JobRepository) *Handler {
	return &Handler{
		LLMGateway: llmGateway,
		JobRepo:    jobRepo,
	}
}

type SelectionRequest struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

func (h *Handler) HandleJob(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/job/")
	id = strings.TrimSpace(id)
	if id == "" {
		http.NotFound(w, r)
		return
	}

	jobID, err := uuid.Parse(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	h.renderJob(w, jobID)
}

func (h *Handler) HandleSelection(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var req SelectionRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	jobMeta := entities.Meta{
		Title: req.Title,
		Text:  req.Text,
	}

	jobID, err := h.JobRepo.CreateJob(jobMeta)
	if err != nil {
		log.Printf("failed to create job: %v", err)
		http.Error(w, "failed to create job", http.StatusInternalServerError)
		return
	}
	go h.processJob(ctx, jobID)

	http.Redirect(w, r, "/job/"+jobID.String(), http.StatusSeeOther)
}

func (h *Handler) processJob(ctx context.Context, jobID uuid.UUID) {
	job := h.JobRepo.GetJob(jobID)
	if job == nil {
		log.Printf("job disappeared before processing: %s", jobID)
		return
	}

	job.Status = entities.StatusRunning
	if err := h.JobRepo.UpdateJob(*job); err != nil {
		log.Printf("failed to mark job %s as running: %v", jobID, err)
		return
	}

	llmRes, err := h.LLMGateway.Generate(ctx, llm.GenerateRequest{
		Prompt: job.Meta.Text,
	})
	if err != nil {
		job.Status = entities.StatusFailed
		job.Error = "Failed to generate recap."
		if updateErr := h.JobRepo.UpdateJob(*job); updateErr != nil {
			log.Printf("failed to mark job %s as failed: %v", jobID, updateErr)
		}
		log.Printf("failed to generate LLM response for job %s: %v", jobID, err)
		return
	}

	job.Summary = llmRes.Text
	job.Status = entities.StatusDone
	if err := h.JobRepo.UpdateJob(*job); err != nil {
		log.Printf("failed to complete job %s: %v", jobID, err)
	}
}

func (h *Handler) renderJob(w http.ResponseWriter, id uuid.UUID) {
	j := h.JobRepo.GetJob(id)
	if j == nil {
		renderPage(w, "Missing recap", "<h1>Missing recap</h1><p>That job id is unknown.</p>", 0)
		return
	}
	safeTitle := html.EscapeString(j.Meta.Title)
	switch j.Status {
	case entities.StatusPending, entities.StatusRunning:
		renderPage(w, "Working...", fmt.Sprintf("<h1>Working...</h1><p>Summarizing <b>%s</b>. This page refreshes automatically.</p>", safeTitle), 3)
	case entities.StatusFailed:
		renderPage(w, "Recap error", fmt.Sprintf("<h1>Could not make recap</h1><div class='box'>%s</div>", html.EscapeString(j.Error)), 0)
	default:
		body := fmt.Sprintf("<h1>Recap: %s</h1><div class='box'>%s</div>", safeTitle, html.EscapeString(j.Summary))
		// body += fmt.Sprintf("<p class='small'>Mode: %s | Input words: %d</p>", html.EscapeString(j.Meta.Type), j.ExcerptWordCount)
		// if j.Meta.Type == "position" {
		// 	body += fmt.Sprintf("<p class='small'>Book words indexed: %d | Position: %d | Chapter match: %s</p>", j.Where.TotalWords, j.Where.EndWord, html.EscapeString(valueOr(j.Where.MatchedChapter, "none")))
		// 	body += fmt.Sprintf("<p class='small'>Book file: %s</p>", html.EscapeString(j.BookPath))
		// }
		renderPage(w, "Kobo recap", body, 0)
	}
}

func renderPage(w http.ResponseWriter, title, body string, refreshSeconds int) {
	refresh := ""
	if refreshSeconds > 0 {
		refresh = fmt.Sprintf(`<meta http-equiv="refresh" content="%d">`, refreshSeconds)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!doctype html>
<html>
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
%s
<title>%s</title>
<style>
body { font-family: Georgia, serif; font-size: 1.15rem; line-height: 1.45; margin: 1rem; max-width: 52rem; }
h1 { font-size: 1.4rem; }
.box { white-space: pre-wrap; border: 1px solid #000; padding: .75rem; }
.small { font-size: .85rem; }
</style>
</head>
<body>
%s
</body>
</html>`, refresh, html.EscapeString(title), body)
}
