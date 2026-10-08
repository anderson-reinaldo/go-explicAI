package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/anderson-reinaldo/go-explicAI/internal/application"
	gatewaymocks "github.com/anderson-reinaldo/go-explicAI/internal/gateway/mocks"
	"github.com/anderson-reinaldo/go-explicAI/internal/gateway/repository"
	"github.com/anderson-reinaldo/go-explicAI/internal/gateway/summarize"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

var (
	timeLayout            = "2006-01-02 15:04:05"
	summaryExternalIDStr  = "4f6a1bc1-c715-4336-a06a-ca465330e01a"
	summaryExternalIDUUID = uuid.MustParse(summaryExternalIDStr)
	createdAt, _          = time.Parse(timeLayout, "2025-01-25 14:04:05")
	textTranscribed       = "result text transcribed"
	title                 = "title"
	description           = "description"
	briefResume           = "brief resume"
	mediumResume          = "medium resume"
	fulltext              = "full text"
)

type (
	SummaryTestSuite struct {
		suite.Suite

		ctx             context.Context
		audiotranscript *gatewaymocks.MockAudioTranscript
		summarize       *gatewaymocks.MockSummarize
		repository      *gatewaymocks.MockReposistory
	}
)

func TestSummaryTestSuite(t *testing.T) {
	suite.Run(t, new(SummaryTestSuite))
}

func (s *SummaryTestSuite) SetupTest() {
	s.ctx = context.Background()
}

func (s *SummaryTestSuite) TearDownTest() {
	mock.AssertExpectationsForObjects(s.T())
}

func (s *SummaryTestSuite) TestSummaryCreate() {
	s.Run("successful create summary", func() {
		s.repository = new(gatewaymocks.MockReposistory)
		s.repository.EXPECT().
			CreateSummary(mock.Anything, repository.ReceivedFile).
			Return(&repository.SummaryCreateOutput{
				ExternalID: summaryExternalIDUUID,
				Status:     repository.StatusToString[repository.ReceivedFile].Status,
				Progress: sql.NullInt32{
					Int32: 33,
					Valid: true,
				},
				CreatedAt: createdAt,
			}, nil)

		s.audiotranscript = new(gatewaymocks.MockAudioTranscript)
		s.audiotranscript.EXPECT().
			Transcribe(mock.Anything, mock.Anything).
			Return(&textTranscribed, nil)

		ustInput := repository.SummaryUpdateSummarizedInput{
			ExternalID: summaryExternalIDUUID,
			Status:     repository.Transcribed,
		}

		s.repository.EXPECT().
			UpdateSummaryTranscribed(mock.Anything, ustInput)

		s.summarize = new(gatewaymocks.MockSummarize)
		s.summarize.EXPECT().
			Resume(mock.Anything, nil).
			Return(&summarize.ResumeOutput{
				Title:        title,
				Description:  description,
				BriefResume:  briefResume,
				MediumResume: mediumResume,
			}, nil)

		s.summarize.EXPECT().
			FullTextOrganize(mock.Anything, textTranscribed).
			Return(&fulltext, nil)

		susInput := repository.SummaryUpdateSummarizedInput{
			ExternalID:   summaryExternalIDUUID,
			Status:       repository.Summarized,
			Title:        title,
			Description:  description,
			BriefResume:  briefResume,
			MediumResume: mediumResume,
			FullText:     fulltext,
		}

		s.repository.EXPECT().
			UpdateSummarySummarized(mock.Anything, susInput).
			Return(nil)

		service := NewSummary(s.audiotranscript, s.summarize, s.repository)
		output, err := service.CreateSummaryAndTriggerAIProcess(s.ctx, []byte{})
		s.Require().NoError(err)
		s.Equal("RECEIVED_FILE", output.Status)
		s.Equal(createdAt, output.CreatedAt)
	})

	s.Run("fail db create summary", func() {
		s.repository = new(gatewaymocks.MockReposistory)
		s.repository.EXPECT().
			CreateSummary(mock.Anything, repository.ReceivedFile).
			Return(nil, errors.New("some error"))

		service := NewSummary(s.audiotranscript, s.summarize, s.repository)
		_, err := service.CreateSummaryAndTriggerAIProcess(s.ctx, []byte{})
		s.Require().ErrorIs(err, application.InternalDatabaseError)
	})
}

func (s *SummaryTestSuite) TestAIProcessSummary() {
	s.Run("sucessful process summary", func() {
		ctx, cancel := context.WithCancel(context.Background())

		s.audiotranscript = new(gatewaymocks.MockAudioTranscript)
		s.audiotranscript.EXPECT().
			Transcribe(mock.Anything, mock.Anything).
			Return(&textTranscribed, nil)

		ustInput := repository.SummaryUpdateTranscribedInput{
			ExternalID: summaryExternalIDUUID,
			Status:     repository.Transcribed,
		}

		s.repository = new(gatewaymocks.MockReposistory)
		s.repository.EXPECT().
			UpdateSummaryTranscribed(mock.Anything, ustInput).
			Return(nil)

		s.summarize = new(gatewaymocks.MockSummarize)
		s.summarize.EXPECT().
			Resume(mock.Anything, textTranscribed).
			Return(&summarize.ResumeOutput{
				Title:        title,
				Description:  description,
				BriefResume:  briefResume,
				MediumResume: mediumResume,
			}, nil)
		s.summarize.EXPECT().
			FullTextOrganize(mock.Anything, textTranscribed).
			Return(&fulltext, nil)

		susInput := repository.SummaryUpdateSummarizedInput{
			ExternalID:   summaryExternalIDUUID,
			Status:       repository.Summarized,
			Title:        title,
			Description:  description,
			BriefResume:  briefResume,
			MediumResume: mediumResume,
			FullText:     fulltext,
		}

		s.repository.EXPECT().
			UpdateSummarySummarized(mock.Anything, susInput).
			Return(nil)

		service := NewSummary(s.audiotranscript, s.summarize, s.repository)
		service.AISummaryProccess(ctx, cancel, []byte{}, summaryExternalIDUUID)

		s.audiotranscript.AssertCalled(s.T(), "Transcribe", mock.Anything, mock.Anything)
		s.repository.AssertNotCalled(s.T(), "UpdateSummaryTranscribed", mock.Anything, mock.Anything)
		s.summarize.AssertCalled(s.T(), "Resume", mock.Anything, textTranscribed)
		s.summarize.AssertCalled(s.T(), "FullTextOrganize", mock.Anything, textTranscribed)
		s.repository.AssertCalled(s.T(), "UpdateSummarySummarized", mock.Anything, susInput)
	})

	s.Run("fail fulltext organize summary", func() {
		ctx, cancel := context.WithCancel(context.Background())

		s.audiotranscript = new(gatewaymocks.MockAudioTranscript)
		s.audiotranscript.EXPECT().
			Transcribe(mock.Anything, mock.Anything).
			Return(&textTranscribed, nil)

		ustInput := repository.SummaryUpdateTranscribedInput{
			ExternalID: summaryExternalIDUUID,
			Status:     repository.Transcribed,
		}

		s.repository = new(gatewaymocks.MockReposistory)
		s.repository.EXPECT().
			UpdateSummaryTranscribed(mock.Anything, ustInput).
			Return(nil)

		s.summarize = new(gatewaymocks.MockSummarize)
		s.summarize.EXPECT().
			Resume(mock.Anything, textTranscribed).
			Return(&summarize.ResumeOutput{
				Title:        title,
				Description:  description,
				BriefResume:  briefResume,
				MediumResume: mediumResume,
			}, nil)
		s.summarize.EXPECT().
			FullTextOrganize(mock.Anything, textTranscribed).
			Return(&fulltext, errors.New("some error"))

		susInput := repository.SummaryUpdateSummarizedInput{
			ExternalID: summaryExternalIDUUID,
			Status:     repository.SummarizedFailed,
		}

		s.repository.EXPECT().
			UpdateSummarySummarized(mock.Anything, susInput).
			Return(nil)

		service := NewSummary(s.audiotranscript, s.summarize, s.repository)
		service.AISummaryProccess(ctx, cancel, []byte{}, summaryExternalIDUUID)

		s.audiotranscript.AssertCalled(s.T(), "Transcribe", mock.Anything, mock.Anything)
		s.repository.AssertNotCalled(s.T(), "UpdateSummaryTranscribed", mock.Anything, mock.Anything)
		s.summarize.AssertCalled(s.T(), "Resume", mock.Anything, textTranscribed)
		s.summarize.AssertCalled(s.T(), "FullTextOrganize", mock.Anything, textTranscribed)
		s.repository.AssertCalled(s.T(), "UpdateSummarySummarized", mock.Anything, susInput)
	})

	s.Run("fail resume summary", func() {
		ctx, cancel := context.WithCancel(context.Background())

		s.audiotranscript = new(gatewaymocks.MockAudioTranscript)
		s.audiotranscript.EXPECT().
			Transcribe(mock.Anything, mock.Anything).
			Return(&textTranscribed, nil)

		ustInput := repository.SummaryUpdateTranscribedInput{
			ExternalID: summaryExternalIDUUID,
			Status:     repository.Transcribed,
		}

		s.repository = new(gatewaymocks.MockReposistory)
		s.repository.EXPECT().
			UpdateSummaryTranscribed(mock.Anything, ustInput).
			Return(nil)

		s.summarize = new(gatewaymocks.MockSummarize)
		s.summarize.EXPECT().
			Resume(mock.Anything, textTranscribed).
			Return(nil, errors.New("some error"))
		s.summarize.EXPECT().
			FullTextOrganize(mock.Anything, textTranscribed).
			Return(&fulltext, nil)

		susInput := repository.SummaryUpdateSummarizedInput{
			ExternalID: summaryExternalIDUUID,
			Status:     repository.SummarizedFailed,
		}

		s.repository.EXPECT().
			UpdateSummarySummarized(mock.Anything, susInput).
			Return(nil)

		service := NewSummary(s.audiotranscript, s.summarize, s.repository)
		service.AISummaryProccess(ctx, cancel, []byte{}, summaryExternalIDUUID)

		s.audiotranscript.AssertCalled(s.T(), "Transcribe", mock.Anything, mock.Anything)
		s.repository.AssertNotCalled(s.T(), "UpdateSummaryTranscribed", mock.Anything, mock.Anything)
		s.summarize.AssertCalled(s.T(), "Resume", mock.Anything, textTranscribed)
		s.summarize.AssertCalled(s.T(), "FullTextOrganize", mock.Anything, textTranscribed)
		s.repository.AssertCalled(s.T(), "UpdateSummarySummarized", mock.Anything, susInput)
	})

	s.Run("fail audio transcribe", func() {
		ctx, cancel := context.WithCancel(context.Background())

		s.audiotranscript = new(gatewaymocks.MockAudioTranscript)
		s.audiotranscript.EXPECT().
			Transcribe(mock.Anything, mock.Anything).
			Return(nil, errors.New("some error"))

		ustInput := repository.SummaryUpdateTranscribedInput{
			ExternalID: summaryExternalIDUUID,
			Status:     repository.TranscribedFailed,
		}

		s.repository = new(gatewaymocks.MockReposistory)
		s.repository.EXPECT().
			UpdateSummaryTranscribed(mock.Anything, ustInput).
			Return(nil)

		service := NewSummary(s.audiotranscript, s.summarize, s.repository)
		service.AISummaryProccess(ctx, cancel, []byte{}, summaryExternalIDUUID)

		s.audiotranscript.AssertCalled(s.T(), "Transcribe", mock.Anything, mock.Anything)
		s.repository.AssertCalled(s.T(), "UpdateSummaryTranscribed", mock.Anything, mock.Anything)
	})

}
