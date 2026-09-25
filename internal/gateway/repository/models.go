package repository

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Status int32

const (
	ReceivedFile Status = iota
	Transcribed
	TranscribedFailed
	Summarized
	SummarizedFailed
)

type StatusDomain struct {
	Status     string
	Percentage int
}

var (
	StatusToString = map[Status]StatusDomain{
		ReceivedFile:      {"RECEIVED_FILE", 33},
		TranscribedFailed: {"TRANSCRIBED_FAILED", 33},
		Transcribed:       {"TRANSCRIBED", 66},
		SummarizedFailed:  {"SUMMARIZED_FAILED", 66},
		Summarized:        {"SUMMARIZED", 100},
	}
)

type (
	SummaryCreateOutput struct {
		ExternalID uuid.UUID
		Status     string
		Progress   sql.NullInt32
		CreatedAt  time.Time
	}
)

type (
	SummaryUpdateSummarizedInput struct {
		ExternalID   uuid.UUID
		Status       Status
		Title        string
		Description  string
		BriefResume  string
		MediumResume string
		FullText     string
	}

	SummaryUpdateTranscribedInput struct {
		ExternalID uuid.UUID
		Status     Status
	}
)

type (
	SummaryOutput struct {
		ExternalID   uuid.UUID
		CreatedAt    time.Time
		UpdateAt     time.Time
		Status       string
		Title        sql.NullString
		Description  sql.NullString
		BriefResume  sql.NullString
		MediumResume sql.NullString
		Progress     sql.NullInt32
		FullText     sql.NullString
	}
)
