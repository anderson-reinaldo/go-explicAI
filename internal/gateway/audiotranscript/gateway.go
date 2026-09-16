package audiotranscript

import "context"

type AudioTranscript interface {
	Transcribe(ctx context.Context, audio []byte) (*string, error)
}
