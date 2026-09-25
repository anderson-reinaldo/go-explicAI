package db

import (
	"context"
	"testing"
	"time"

	"github.com/anderson-reinaldo/go-explicAI/internal/gateway/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	postgresDockerImage = "postgres:16-alpine"
)

type SummaryDBTestSuite struct {
	suite.Suite
	ctx         context.Context
	container   *postgres.PostgresContainer
	databaseURL string
	summaryDB   *Summary
}

func TestSummaryDB(t *testing.T) {
	suite.Run(t, new(SummaryDBTestSuite))
}

func (s *SummaryDBTestSuite) SetupSuite() {
	var err error
	s.ctx = context.Background()

	s.container, err = postgres.Run(s.ctx,
		postgresDockerImage,
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connection").
				WithOccurrence(2).WithStartupTimeout(5*time.Second),
		),
	)

	s.NoError(err)

	s.databaseURL, err = s.container.ConnectionString(s.ctx)
	s.NoError(err)

	s.summaryDB = NewSummary(s.databaseURL)

	conn := NewPgConnection(s.databaseURL)
	pgConn, err := conn.Connect(s.ctx)
	s.NoError(err)

	defer conn.Close(s.ctx, pgConn)

	_, err = pgConn.Exec(s.ctx, `
	CREATE TABLE summaries (
		id SERIAL PRIMARY KEY,
		external_id UUID NOT NULL,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL,
		status VARCHAR(50) NOT NULL,
		title VARCHAR(255),
		description TEXT,
		brief_resume TEXT,
		medium_resume TEXT,
		progress int,
		fulltext TEXT
		);
	`)
	s.NoError(err)
}

func (s *SummaryDBTestSuite) TearDownSuite() {
	s.NoError(s.container.Terminate(s.ctx))
}

func (s *SummaryDBTestSuite) TestSummaryDBOperations() {
	s.Run("successful creation of summary", func() {
		output, err := s.summaryDB.CreateSummary(s.ctx, repository.ReceivedFile)
		s.NoError(err)
		s.NotNil(output)
		s.NotEmpty(output.ExternalID)
		s.Equal("RECEIVED_FILE", output.Status)
		s.Equal(33, int(output.Progress.Int32))
	})

	s.Run("successful update of summary progress", func() {
		output, err := s.summaryDB.CreateSummary(s.ctx, repository.ReceivedFile)
		s.NoError(err)

		err = s.summaryDB.UpdateSummaryTranscribed(s.ctx, repository.SummaryUpdateTranscribedInput{
			ExternalID: output.ExternalID,
			Status:     repository.Transcribed,
		})

		s.NoError(err)

		result, err := s.summaryDB.GetSummaryByExternalID(s.ctx, output.ExternalID)
		s.NoError(err)
		s.Equal("TRANSCRIBED", result.Status)
		s.Equal(66, int(result.Progress.Int32))
	})

	s.Run("successful update of summary progress total", func() {
		output, err := s.summaryDB.CreateSummary(s.ctx, repository.ReceivedFile)
		s.NoError(err)

		err = s.summaryDB.UpdateSummarySummarized(s.ctx, repository.SummaryUpdateSummarizedInput{
			ExternalID:   output.ExternalID,
			Status:       repository.Summarized,
			Title:        "title",
			Description:  "desc",
			BriefResume:  "brief",
			MediumResume: "medium",
			FullText:     "full",
		})

		s.NoError(err)

		result, err := s.summaryDB.GetSummaryByExternalID(s.ctx, output.ExternalID)
		s.NoError(err)
		s.Equal("SUMMARIZED", result.Status)
		s.Equal(100, int(result.Progress.Int32))
		s.Equal("title", result.Title.String)
		s.Equal("desc", result.Description.String)
		s.Equal("brief", result.BriefResume.String)
		s.Equal("medium", result.MediumResume.String)
		s.Equal("full", result.FullText.String)
	})

	s.Run("successful list summaries", func() {
		s.truncate()

		output, err := s.summaryDB.CreateSummary(s.ctx, repository.ReceivedFile)
		s.NoError(err)
		err = s.summaryDB.UpdateSummarySummarized(s.ctx,
			repository.SummaryUpdateSummarizedInput{
				ExternalID:   output.ExternalID,
				Status:       repository.Summarized,
				Title:        "title",
				Description:  "desc",
				BriefResume:  "brief",
				MediumResume: "medium",
				FullText:     "full",
			},
		)
		s.NoError(err)

		s.summaryDB.CreateSummary(s.ctx, repository.ReceivedFile)
		s.summaryDB.CreateSummary(s.ctx, repository.ReceivedFile)
		s.summaryDB.CreateSummary(s.ctx, repository.ReceivedFile)
		s.summaryDB.CreateSummary(s.ctx, repository.ReceivedFile)

		results, err := s.summaryDB.GetSummaries(s.ctx)

		result := results[4]

		s.NoError(err)
		s.Equal(5, len(results))
		s.Equal("SUMMARIZED", result.Status)
		s.Equal(100, int(result.Progress.Int32))
		s.Equal("title", result.Title.String)
		s.Equal("desc", result.Description.String)
		s.Equal("brief", result.BriefResume.String)
		s.Equal("medium", result.MediumResume.String)
		s.Equal("full", result.FullText.String)
	})

	s.Run("successful delete summaries", func() {
		s.truncate()

		s.summaryDB.CreateSummary(s.ctx, repository.ReceivedFile)
		s.summaryDB.CreateSummary(s.ctx, repository.ReceivedFile)
		s.summaryDB.CreateSummary(s.ctx, repository.ReceivedFile)
		targetDelete1, _ := s.summaryDB.CreateSummary(s.ctx, repository.ReceivedFile)
		targetDelete2, _ := s.summaryDB.CreateSummary(s.ctx, repository.ReceivedFile)

		s.summaryDB.DeleteSummaryByExternalID(s.ctx, targetDelete1.ExternalID)
		s.summaryDB.DeleteSummaryByExternalID(s.ctx, targetDelete2.ExternalID)

		results, err := s.summaryDB.GetSummaries(s.ctx)
		s.NoError(err)

		s.Equal(3, len(results))
	})

	s.Run("database connection failure", func() {
		faultyDatabaseURL := "postgresql://invalid_user:invalid_pass@localhost5432/invalid_database"
		faultySummaryDB := NewSummary(faultyDatabaseURL)

		_, err := faultySummaryDB.CreateSummary(s.ctx, repository.ReceivedFile)
		s.Error(err)
	})

	s.Run("update non-existing summary", func() {
		nonExistentID := uuid.New()

		err := s.summaryDB.UpdateSummaryTranscribed(s.ctx,
			repository.SummaryUpdateTranscribedInput{
				ExternalID: nonExistentID,
				Status:     repository.Transcribed,
			})

		s.Error(err)
		s.Contains(err.Error(), "register not found")
	})

	s.Run("delete non-existing summary", func() {
		nonExistentID := uuid.New()

		err := s.summaryDB.DeleteSummaryByExternalID(s.ctx, nonExistentID)

		s.Error(err)
		s.Contains(err.Error(), "no rows affected.")
	})
}

func (s *SummaryDBTestSuite) truncate() {
	conn := NewPgConnection(s.databaseURL)
	pgConn, err := conn.Connect(s.ctx)
	s.NoError(err)
	defer conn.Close(s.ctx, pgConn)

	_, err = pgConn.Exec(s.ctx, "truncate summaries restart identity cascade")
	s.NoError(err)
}
