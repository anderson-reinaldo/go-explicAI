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

func (s *SummaryTestSuite) TestListSummaries() {
	s.Run("successful list summaries", func() {
		summaryExternalID2Str := "1032d178-28c0-43cf-9dda-8b99c6ff3838"
		summaryExternalID2UUID := uuid.MustParse(summaryExternalID2Str)

		s.repository = new(gatewaymocks.MockReposistory)
		s.repository.EXPECT().GetSummaries(mock.Anything).
			Return([]repository.SummaryOutput{
				{
					ExternalID: summaryExternalIDUUID,
					CreatedAt:  createdAt,
					UpdateAt:   createdAt,
					Status:     "SUMMARIZED",
					Title: sql.NullString{
						String: title,
						Valid:  true,
					},
					Description: sql.NullString{
						String: description,
						Valid:  true,
					},
					BriefResume: sql.NullString{
						String: briefResume,
						Valid:  true,
					},
					MediumResume: sql.NullString{
						String: mediumResume,
						Valid:  true,
					},
					FullText: sql.NullString{
						String: fulltext,
						Valid:  true,
					},
					Progress: sql.NullInt32{
						Int32: 100,
						Valid: true,
					},
				},
				{
					ExternalID: summaryExternalID2UUID,
					CreatedAt:  createdAt,
					Status:     "RECEIVED_FILE",
				},
			}, nil)

		service := NewSummary(s.audiotranscript, s.summarize, s.repository)
		output, err := service.ListSummaries(s.ctx)
		s.Require().NoError(err)
		s.Equal(summaryExternalIDUUID, output.Data[0].ExternalID)
		s.Equal("SUMMARIZED", output.Data[0].Status)
		s.Equal(createdAt, output.Data[0].CreatedAt)
		s.Equal(createdAt, output.Data[0].UpdatedAt)
		s.Equal(title, output.Data[0].Title)
		s.Equal(description, output.Data[0].Description)
		s.Equal(100, output.Data[0].Progress)

		s.Equal(summaryExternalID2UUID, output.Data[1].ExternalID)
		s.Equal("RECEIVED_FILE", output.Data[1].Status)
		s.Equal(createdAt, output.Data[1].CreatedAt)
	})

	s.Run("fail list summaries", func() {
		s.repository = new(gatewaymocks.MockReposistory)
		s.repository.EXPECT().GetSummaries(s.ctx).
			Return(nil, errors.New("some error"))

		service := NewSummary(s.audiotranscript, s.summarize, s.repository)
		_, err := service.ListSummaries(s.ctx)
		s.Require().ErrorIs(err, application.UnexpectedErrorList)
	})
}

func (s *SummaryTestSuite) TestGetSummaryByExternalID() {
	s.Run("successful get summary by external id", func() {
		s.repository = new(gatewaymocks.MockReposistory)
		s.repository.EXPECT().
			GetSummaryByExternalID(mock.Anything, summaryExternalIDUUID).
			Return(&repository.SummaryOutput{
				ExternalID: summaryExternalIDUUID,
				Status:     repository.StatusToString[repository.Summarized].Status,
				CreatedAt:  createdAt,
				UpdateAt:   createdAt,
				Title: sql.NullString{
					String: title,
					Valid:  true,
				},
				Description: sql.NullString{
					String: description,
					Valid:  true,
				},
				BriefResume: sql.NullString{
					String: briefResume,
					Valid:  true,
				},
				MediumResume: sql.NullString{
					String: mediumResume,
					Valid:  true,
				},
				FullText: sql.NullString{
					String: fulltext,
					Valid:  true,
				},
				Progress: sql.NullInt32{
					Int32: 100,
					Valid: true,
				},
			}, nil)

		service := NewSummary(s.audiotranscript, s.summarize, s.repository)
		output, err := service.GetSummaryByExternalID(s.ctx, summaryExternalIDUUID)
		s.Require().NoError(err)
		s.Equal(summaryExternalIDUUID, output.ExternalID)
		s.Equal(repository.StatusToString[repository.Summarized].Status, output.Status)
		s.Equal(100, output.Progress)
		s.Equal(title, output.Title)
		s.Equal(description, output.Description)
		s.Equal(briefResume, output.BriefResume)
		s.Equal(mediumResume, output.MediumResume)

	})

	s.Run("error get summary by external id", func() {
		s.repository = new(gatewaymocks.MockReposistory)
		s.repository.EXPECT().
			GetSummaryByExternalID(mock.Anything, summaryExternalIDUUID).
			Return(nil, errors.New("some error"))

		service := NewSummary(s.audiotranscript, s.summarize, s.repository)
		_, err := service.GetSummaryByExternalID(s.ctx, summaryExternalIDUUID)
		s.Require().Error(err)
	})

	s.Run("get summary by external id not found", func() {
		s.repository = new(gatewaymocks.MockReposistory)
		s.repository.EXPECT().
			GetSummaryByExternalID(mock.Anything, summaryExternalIDUUID).
			Return(nil, application.SummaryNotFound)

		service := NewSummary(s.audiotranscript, s.summarize, s.repository)
		_, err := service.GetSummaryByExternalID(s.ctx, summaryExternalIDUUID)
		s.Require().ErrorIs(err, application.SummaryNotFound)
	})
}

func (s *SummaryTestSuite) TestDeleteSummaryByExternalID() {
	s.Run("successful delete summary by external id", func() {
		s.repository = new(gatewaymocks.MockReposistory)
		s.repository.EXPECT().
			DeleteSummaryByExternalID(mock.Anything, summaryExternalIDUUID).
			Return(nil)

		service := NewSummary(s.audiotranscript, s.summarize, s.repository)
		err := service.DeleteSummaryByExternalID(s.ctx, summaryExternalIDUUID)
		s.Require().NoError(err)
	})

	s.Run("error delete summary by external id", func() {
		s.repository = new(gatewaymocks.MockReposistory)
		s.repository.EXPECT().
			DeleteSummaryByExternalID(mock.Anything, summaryExternalIDUUID).
			Return(errors.New("some error"))

		service := NewSummary(s.audiotranscript, s.summarize, s.repository)
		err := service.DeleteSummaryByExternalID(s.ctx, summaryExternalIDUUID)
		s.Require().Error(err)
	})

	s.Run("delete summary by external id not found", func() {
		s.repository = new(gatewaymocks.MockReposistory)
		s.repository.EXPECT().
			DeleteSummaryByExternalID(mock.Anything, summaryExternalIDUUID).
			Return(application.SummaryNotFound)

		service := NewSummary(s.audiotranscript, s.summarize, s.repository)
		err := service.DeleteSummaryByExternalID(s.ctx, summaryExternalIDUUID)
		s.Require().ErrorIs(err, application.SummaryNotFound)
	})
}
