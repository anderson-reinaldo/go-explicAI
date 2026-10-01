package service

import (
	"time"

	"github.com/google/uuid"
)

type (
	SummarySimpleOutput struct {
		ExternalID  uuid.UUID `json:"externalId"`
		Status      string    `json:"status"`
		CreatedAt   time.Time `json:"createdAt"`
		UpdatedAt   time.Time `json:"updatedAt"`
		Progress    int       `json:"progress"`
		Title       string    `json:"title,omitempty"`
		Description string    `json:"description,omitempty"`
	}

	SummaryDetailedOutput struct {
		ExternalID   uuid.UUID `json:"externalId"`
		Status       string    `json:"status"`
		CreatedAt    time.Time `json:"createdAt"`
		UpdatedAt    time.Time `json:"updatedAt"`
		Progress     int       `json:"progress"`
		Title        string    `json:"title,omitempty"`
		Description  string    `json:"description,omitempty"`
		BriefResume  string    `json:"briefResume,omitempty"`
		MediumResume string    `json:"mediumResume,omitempty"`
		FullText     string    `json:"fullText,omitempty"`
	}

	SummaryListOutput struct {
		Data []SummarySimpleOutput `json:"data"`
	}
)
