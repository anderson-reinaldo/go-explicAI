package service

import (
	"context"

	"github.com/anderson-reinaldo/go-explicAI/internal/application"
	"github.com/anderson-reinaldo/go-explicAI/internal/gateway/audiotranscript"
	"github.com/anderson-reinaldo/go-explicAI/internal/gateway/repository"
	"github.com/anderson-reinaldo/go-explicAI/internal/gateway/summarize"
	"github.com/anderson-reinaldo/go-explicAI/internal/infrastructure/log"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type SummaryUseCase interface {
	CreateSummaryAndTriggerAIProcess(ctx context.Context, audio []byte) (*SummarySimpleOutput, error)
	ListSummaries(ctx context.Context) (*SummaryListOutput, error)
	GetSummaryByExternalID(ctx context.Context, externalID uuid.UUID) (*SummaryDetailedOutput, error)
	DeleteSummaryByExternalID(ctx context.Context, externalID uuid.UUID) error
}

type Summary struct {
	audiotranscript audiotranscript.AudioTranscript
	sumarize        summarize.Summarize
	repository      repository.Reposistory
}

func NewSummary(
	audiotranscript audiotranscript.AudioTranscript,
	sumarize summarize.Summarize,
	repository repository.Reposistory,
) *Summary {
	return &Summary{
		audiotranscript: audiotranscript,
		sumarize:        sumarize,
		repository:      repository,
	}
}

func (s *Summary) CreateSummaryAndTriggerAIProcess(ctx context.Context, audio []byte) (*SummarySimpleOutput, error) {
	r, err := s.repository.CreateSummary(ctx, repository.ReceivedFile)
	if err != nil {
		log.LogError(ctx, "failed to create summary in db", err)
		return nil, application.InternalDatabaseError
	}

	ctx, cancel := context.WithCancel(context.Background())

	go s.AISummaryProccess(ctx, cancel, audio, r.ExternalID)

	progress := repository.StatusToString[repository.ReceivedFile].Percentage

	if r.Progress.Valid {
		progress = int(r.Progress.Int32)
	}

	return &SummarySimpleOutput{
		ExternalID: r.ExternalID,
		Status:     r.Status,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.CreatedAt,
		Progress:   progress,
	}, nil

}

func (s *Summary) ListSummaries(ctx context.Context) (*SummaryListOutput, error) {
	summaries, err := s.repository.GetSummaries(ctx)
	if err != nil {
		return nil, application.UnexpectedErrorList
	}

	var summariesOutput []SummarySimpleOutput

	for _, sum := range summaries {
		summariesOutput = append(summariesOutput, SummarySimpleOutput{
			ExternalID:  sum.ExternalID,
			Status:      sum.Status,
			CreatedAt:   sum.CreatedAt,
			UpdatedAt:   sum.UpdateAt,
			Progress:    int(sum.Progress.Int32),
			Title:       sum.Title.String,
			Description: sum.Description.String,
		})

	}

	return &SummaryListOutput{
		Data: summariesOutput,
	}, nil
}

func (s *Summary) GetSummaryByExternalID(ctx context.Context, externalID uuid.UUID) (*SummaryDetailedOutput, error) {
	summary, err := s.repository.GetSummaryByExternalID(ctx, externalID)
	if err == application.SummaryNotFound {
		return nil, err
	}

	if err != nil {
		log.LogError(ctx, "error on get summary", err)
		return nil, err
	}

	return &SummaryDetailedOutput{
		ExternalID:   summary.ExternalID,
		Status:       summary.Status,
		CreatedAt:    summary.CreatedAt,
		UpdatedAt:    summary.UpdateAt,
		Progress:     int(summary.Progress.Int32),
		Title:        summary.Title.String,
		Description:  summary.Description.String,
		BriefResume:  summary.BriefResume.String,
		MediumResume: summary.MediumResume.String,
		FullText:     summary.FullText.String,
	}, nil
}

func (s *Summary) DeleteSummaryByExternalID(ctx context.Context, externalID uuid.UUID) error {
	err := s.repository.DeleteSummaryByExternalID(ctx, externalID)
	if err == application.SummaryNotFound {
		return err
	}

	if err != nil {
		log.LogError(ctx, "error on delete summary", err)
		return err
	}
	return nil
}

func (s *Summary) AISummaryProccess(ctx context.Context, cancel context.CancelFunc, audio []byte, externalID uuid.UUID) {
	defer cancel()

	transcribe, err := s.AudioTranscribe(ctx, audio, externalID)
	if err != nil {
		return
	}

	var resume summarize.ResumeOutput
	var fulltext string

	g := new(errgroup.Group)

	g.Go(func() error {
		return s.resumeText(ctx, *transcribe, externalID, &resume)
	})

	g.Go(func() error {
		return s.organizeText(ctx, *transcribe, externalID, &fulltext)
	})

	if g.Wait() != nil {
		s.registerSummarizedFailed(ctx, externalID)
		return
	}

	s.registerSummarizedSuccess(ctx, externalID, resume, fulltext)

}

func (s *Summary) AudioTranscribe(ctx context.Context, audio []byte, externalID uuid.UUID) (*string, error) {
	log.LogInfo(ctx, "start audio transcribe", zap.String("external_id", externalID.String()))

	transcription, err := s.audiotranscript.Transcribe(ctx, audio)
	if err != nil {
		log.LogError(ctx, "failed to transcript text", err, zap.String("external_id", externalID.String()))
		s.registerTranscribeFailed(ctx, externalID)
		return nil, err
	}

	return transcription, nil

}

func (s *Summary) registerTranscribeFailed(ctx context.Context, externalID uuid.UUID) {
	if err := s.repository.UpdateSummaryTranscribed(ctx, repository.SummaryUpdateTranscribedInput{
		ExternalID: externalID,
		Status:     repository.TranscribedFailed,
	}); err != nil {
		log.LogError(ctx, "failed to save in db", err)
	}
}

func (s *Summary) resumeText(ctx context.Context, transcription string, externalID uuid.UUID, response *summarize.ResumeOutput) error {
	log.LogInfo(ctx, "start resume transcription", zap.String("external_id", externalID.String()))

	result, err := s.sumarize.Resume(ctx, transcription)
	if err != nil {
		log.LogError(ctx, "failed resume transcription", err, zap.String("external_id", externalID.String()))
		return application.ResumeTextFailed
	}

	*response = *result
	log.LogInfo(ctx, "successful resume transcription", zap.String("external_id", externalID.String()))
	return nil
}

func (s *Summary) organizeText(ctx context.Context, transcription string, externalID uuid.UUID, response *string) error {
	log.LogInfo(ctx, "start full organize text", zap.String("external_id", externalID.String()))

	result, err := s.sumarize.FullTextOrganize(ctx, transcription)
	if err != nil {
		log.LogError(ctx, "failed full organize text", err, zap.String("external_id", externalID.String()))
		return application.ResumeTextFailed
	}

	*response = *result
	log.LogInfo(ctx, "successful full organize text", zap.String("external_id", externalID.String()))
	return nil
}

func (s *Summary) registerSummarizedSuccess(ctx context.Context, externalID uuid.UUID, resume summarize.ResumeOutput, fulltext string) {
	if err := s.repository.UpdateSummarySummarized(ctx, repository.SummaryUpdateSummarizedInput{
		ExternalID:   externalID,
		Status:       repository.Summarized,
		Title:        resume.Title,
		Description:  resume.Description,
		BriefResume:  resume.BriefResume,
		MediumResume: resume.MediumResume,
		FullText:     fulltext,
	}); err != nil {
		log.LogError(ctx, "failed to save in db", err)
	}
}

func (s *Summary) registerSummarizedFailed(ctx context.Context, externalID uuid.UUID) {
	if err := s.repository.UpdateSummarySummarized(ctx, repository.SummaryUpdateSummarizedInput{
		ExternalID: externalID,
		Status:     repository.SummarizedFailed,
	}); err != nil {
		log.LogError(ctx, "failed to save in db", err)
	}
}
