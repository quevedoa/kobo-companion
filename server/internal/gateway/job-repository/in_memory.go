package jobrepository

import (
	"fmt"
	"kobo-companion/internal/entities"
	"sync"

	"github.com/google/uuid"
)

type InMemoryJobRepo struct {
	mu     sync.RWMutex
	Memo   map[uuid.UUID]*entities.Job
	Latest *entities.Job
}

func NewInMemoryJobRepo() *InMemoryJobRepo {
	return &InMemoryJobRepo{
		Memo: map[uuid.UUID]*entities.Job{},
	}
}

func (r *InMemoryJobRepo) GetLatest() *entities.Job {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.Latest
}

func (r *InMemoryJobRepo) CreateJob(meta entities.Meta) (uuid.UUID, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	jobID := uuid.New()
	job := &entities.Job{
		ID:     jobID,
		Meta:   meta,
		Status: entities.StatusPending,
	}
	r.Memo[jobID] = job
	r.Latest = job
	return jobID, nil
}

func (r *InMemoryJobRepo) UpdateJob(job entities.Job) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.Memo[job.ID]; ok {
		r.Memo[job.ID] = &job
		return nil
	}
	return fmt.Errorf("job ID not in repo: %s", job.ID)
}

func (r *InMemoryJobRepo) GetJob(jobID uuid.UUID) *entities.Job {
	r.mu.RLock()
	defer r.mu.RUnlock()

	job := r.Memo[jobID]
	if job == nil {
		return nil
	}

	jobCopy := *job
	return &jobCopy
}
