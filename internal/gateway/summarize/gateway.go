package summarize

import "context"

type Summarize interface {
	resume(ctx context.Context, transcription string) (*ResumeOutput, error)
	FullTextOrganize(ctx context.Context, trascription string) (*string, error)
}
