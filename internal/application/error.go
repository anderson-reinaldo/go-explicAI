package application

import "errors"

var (
	MissingFile           = errors.New("missing file to upload")
	InvalidFile           = errors.New("invalid file")
	FailedReadFile        = errors.New("fail read to upload")
	SummaryNotFound       = errors.New("summary not found")
	ExternalIDIsInvalid   = errors.New("externalID is invalid")
	InternalDatabaseError = errors.New("internal database error")
	ResumeTextFailed      = errors.New("resume text failed")
)
