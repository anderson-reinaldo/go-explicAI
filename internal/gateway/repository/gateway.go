package repository

import (
	"context"

	"github.com/google/uuid"
)

type Reposistory interface {
	CreateSummary(ctx context.Context, status Status) (*SummaryCreateOutput, error)
	UpdateSummaryTranscribed(ctx context.Context, input SummaryUpdateTranscribedInput) error
	UpdateSummarySummarized(ctx context.Context, input SummaryUpdateSummarizedInput) error
	GetSummaries(ctx context.Context) ([]SummaryOutput, error)
	GetSummaryByExternalID(ctx context.Context, externalID uuid.UUID) (*SummaryOutput, error)
	DeleteSummaryByExternalID(ctx context.Context, externalID uuid.UUID) error
}
