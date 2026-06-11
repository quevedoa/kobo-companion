package jobrepository

import (
	"kobo-companion/internal/entities"

	"github.com/google/uuid"
)

type JobRepository interface {
	CreateJob(meta entities.Meta) (uuid.UUID, error)
	UpdateJob(job entities.Job) error
	GetJob(jobID uuid.UUID) *entities.Job
	GetLatest() *entities.Job
}
